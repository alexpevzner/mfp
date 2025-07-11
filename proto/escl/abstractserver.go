// MFP - Miulti-Function Printers and scanners toolkit
// eSCL core protocol
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// eSCL server on a top of abstract.Scanner

package escl

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/OpenPrinting/go-mfp/abstract"
	"github.com/OpenPrinting/go-mfp/log"
	"github.com/OpenPrinting/go-mfp/transport"
	"github.com/OpenPrinting/go-mfp/util/missed"
	"github.com/OpenPrinting/go-mfp/util/optional"
	"github.com/OpenPrinting/go-mfp/util/uuid"
	"github.com/OpenPrinting/go-mfp/util/xmldoc"
)

// AbstractServerHistorySize specifies how many scan jobs the
// [AbstractServer] keeps on its history.
const AbstractServerHistorySize = 10

// AbstractServer implements eSCL server on a top of [abstract.Scanner].
type AbstractServer struct {
	ctx      context.Context               // Logging context
	options  AbstractServerOptions         // Server options
	caps     *abstract.ScannerCapabilities // Scanner capabilities
	status   ScannerStatus                 // Scanner status
	document abstract.Document             // Document being server
	lock     sync.Mutex                    // Access lock
}

// AbstractServerOptions represents the [AbstractServerOptions]
// creation options.
type AbstractServerOptions struct {
	Version Version          // eSCL version, DefaultVersion, if not set
	Scanner abstract.Scanner // Underlying abstract.Scanner

	// The BasePath parameter is required so server knows how to
	// interpret [url.URL.Path] of the incoming requests.
	//
	// For the standard eSCL server that mimics the behavior of the
	// typical hardware eSCL scanner, the URL should be something like
	// "/eSCL".
	BasePath string
}

// abstractServerQuery maintains an AbstractServer query processing
// context, allowing per-request centralized logging and hooking.
//
// It keeps the reference to the original [http.Request] and wraps
// the corresponding [http.ResponseWriter], passed to the
// AbstractServer.ServeHTTP
type abstractServerQuery struct {
	log                 *log.Record  // Log record for the query
	*http.Request                    // Incoming request
	http.ResponseWriter              // Underlying http.ResponseWriter
	status              atomic.Int32 // HTTP status, 0 if not known yet
}

// newAbstractServerQuery returns the new abstractServerQuery
func newAbstractServerQuery(srv *AbstractServer,
	w http.ResponseWriter, rq *http.Request) *abstractServerQuery {

	query := &abstractServerQuery{
		log:            log.Begin(srv.ctx),
		Request:        rq,
		ResponseWriter: w,
	}

	return query
}

// RequestHeader returns http.Header of the request
func (query *abstractServerQuery) RequestHeader() http.Header {
	return query.Request.Header
}

// Finish must be called when query processing is finished
func (query *abstractServerQuery) Finish() {
	query.log.Commit()
}

// RequestBody returns body of the http.Request
func (query *abstractServerQuery) RequestBody() io.ReadCloser {
	return query.Request.Body
}

// ResponseHeader returns http.Header of the response
func (query *abstractServerQuery) ResponseHeader() http.Header {
	return query.ResponseWriter.Header()
}

// Write writes response body bytes.
func (query *abstractServerQuery) Write(data []byte) (int, error) {
	return query.ResponseWriter.Write(data)
}

// WriteHeader writes HTTP response header.
func (query *abstractServerQuery) WriteHeader(status int) {
	if query.status.CompareAndSwap(0, int32(status)) {
		query.ResponseWriter.WriteHeader(status)
		query.log.Debug("HTTP %s %s -- %d %s",
			query.Method, query.URL,
			status, http.StatusText(status))
		query.log.Flush()
	}
}

// NoCache set response headers to disable client-side response cacheing.
func (query *abstractServerQuery) NoCache() {
	hdr := query.ResponseHeader()
	hdr.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	hdr.Set("Pragma", "no-cache")
	hdr.Set("Expires", "0")
}

// Reject completes request with a error.
func (query *abstractServerQuery) Reject(status int, err error) {
	query.ResponseHeader().Set("Content-Type", "text/plain; charset=utf-8")
	query.NoCache()
	query.WriteHeader(status)

	if err == nil {
		err = errors.New(http.StatusText(status))
	}

	s := fmt.Sprintf("%3.3d %s\n", status, err)
	query.Write([]byte(s))
	query.Write([]byte("\n"))
}

// Created completes request with the http.StatusCreated
// status and Location: URL
func (query *abstractServerQuery) Created(joburi string) {
	scheme := "http"
	if query.TLS != nil {
		scheme = "https"
	}

	location := scheme + "://" + query.Host + joburi

	query.ResponseHeader().Set("Location", location)
	query.WriteHeader(http.StatusCreated)
}

// SendXML sends the XML response.
func (query *abstractServerQuery) SendXML(xml xmldoc.Element) {
	query.ResponseHeader().Set("Content-Type", HTTPContentType)
	query.WriteHeader(http.StatusOK)
	xml.EncodeIndent(query, NsMap, "  ")
}

// SendImage sends the scanned image
func (query *abstractServerQuery) SendImage(file abstract.DocumentFile) {
	query.ResponseHeader().Set("Content-Type", file.Format())
	query.WriteHeader(http.StatusOK)
	io.Copy(query, file)
}

// NewAbstractServer returns a new [AbstractServer].
func NewAbstractServer(ctx context.Context,
	options AbstractServerOptions) *AbstractServer {

	// Use DefaultVersion, if options.Version is not set
	if options.Version == 0 {
		options.Version = DefaultVersion
	}

	// Canonicalize the base path
	options.BasePath = transport.CleanURLPath(options.BasePath + "/")

	// Create the AbstractServer structure
	srv := &AbstractServer{
		ctx:     ctx,
		options: options,
		caps:    options.Scanner.Capabilities(),
	}

	srv.status = ScannerStatus{
		Version: options.Version,
		State:   ScannerIdle,
	}

	if srv.caps.ADFSimplex != nil || srv.caps.ADFDuplex != nil {
		srv.status.ADFState = optional.New(ScannerAdfProcessing)
	}

	return srv
}

// ServeHTTP serves incoming HTTP requests.
// It implements the [http.Handler] interface.
func (srv *AbstractServer) ServeHTTP(w http.ResponseWriter, rq *http.Request) {
	// Create a abstractServerQuery
	query := newAbstractServerQuery(srv, w, rq)
	defer query.Finish()

	// Dispatch the request
	if !strings.HasPrefix(query.URL.Path, srv.options.BasePath) {
		query.Reject(http.StatusNotFound, nil)
		return
	}

	path, _ := missed.StringsCutPrefix(query.URL.Path,
		srv.options.BasePath)

	// Handle {root}-relative requests
	var action func(*abstractServerQuery)

	srv.lock.Lock()

	switch path {
	case "ScannerCapabilities":
		if query.Method == "GET" {
			action = srv.getScannerCapabilities
		}

	case "ScannerStatus":
		if rq.Method == "GET" {
			action = srv.getScannerStatus
		}

	case "ScanJobs":
		if rq.Method == "POST" {
			action = srv.postScanJobs
		}
	}

	// Handle {JobUri}-relative requests
	if action == nil && srv.document != nil {
		joburi := srv.status.Jobs[0].JobURI

		switch rq.Method {
		case "GET":
			switch query.URL.Path {
			case joburi + "/NextDocument":
				action = srv.getJobURINextDocument
			case joburi + "/ScanImageInfo":
				action = srv.getJobURIScanImageInfo
			}

		case "DELETE":
			if query.URL.Path == joburi {
				action = srv.deleteJobURI
			}
		}
	}

	srv.lock.Unlock()

	if action != nil {
		action(query)
	} else {
		query.Reject(http.StatusNotFound, nil)
	}
}

// getScannerCapabilities handles GET /{root}/ScannerCapabilities request
func (srv *AbstractServer) getScannerCapabilities(query *abstractServerQuery) {
	ver := srv.status.Version
	xml := fromAbstractScannerCapabilities(ver, srv.caps).ToXML()
	query.SendXML(xml)
}

// getScannerStatus handles GET /{root}/ScannerStatus request
func (srv *AbstractServer) getScannerStatus(query *abstractServerQuery) {
	srv.lock.Lock()
	xml := srv.status.ToXML()
	srv.lock.Unlock()

	query.SendXML(xml)
}

// postScanJobs handles POST /{root}/ScanJobs
func (srv *AbstractServer) postScanJobs(query *abstractServerQuery) {
	srv.lock.Lock()
	defer srv.lock.Unlock()

	// Fetch the XML request body
	xml, err := xmldoc.Decode(NsMap, query.RequestBody())
	if err != nil {
		query.Reject(http.StatusBadRequest, err)
		return
	}

	// Decode ScanSettings request
	ss, err := DecodeScanSettings(xml)
	if err != nil {
		query.Reject(http.StatusBadRequest, err)
		return
	}

	// Check if previous request already in progress
	if srv.document != nil {
		err := errors.New("Device is busy with the previous request")
		query.Reject(http.StatusServiceUnavailable, err)
		return
	}

	// Convert it into the abstract.ScannerRequest and validate
	absreq := ss.ToAbstract()

	// Generate a new Job UUID. Do it now, because in theory
	// it can fail (though very unlikely), so do it before
	// the job is created
	uu, err := uuid.Random()
	if err != nil {
		query.Reject(http.StatusServiceUnavailable, err)
		return
	}

	// Send request to the underlying abstract.Scanner
	document, err := srv.options.Scanner.Scan(srv.ctx, absreq)
	if err != nil {
		query.Reject(http.StatusConflict, err)
		return
	}

	// Update server status
	srv.document = document
	srv.status.State = ScannerProcessing

	jobuuid := uu.URN()
	joburi := path.Join(srv.options.BasePath, "ScanJobs", jobuuid)

	info := JobInfo{
		JobURI:   joburi,
		JobUUID:  optional.New(jobuuid),
		JobState: JobProcessing,
	}

	srv.status.PushJobInfo(info, AbstractServerHistorySize)

	// Complete the request
	query.Created(joburi)
}

// getJobURINextDocument handles GET /{JobUri}/NextDocument
func (srv *AbstractServer) getJobURINextDocument(query *abstractServerQuery) {
	srv.lock.Lock()
	file, err := srv.document.Next()
	srv.lock.Unlock()

	switch {
	case err == io.EOF:
		srv.finish(JobCompleted, JobCompletedSuccessfully)
		query.Reject(http.StatusNotFound, nil)

	case err != nil:
		srv.finish(JobCanceled, AbortedBySystem)
		query.Reject(http.StatusServiceUnavailable, err)

	default:
		query.SendImage(file)
	}
}

// getJobURIScanImageInfo handles GET /{JobUri}/ScanImageInfo
func (srv *AbstractServer) getJobURIScanImageInfo(query *abstractServerQuery) {
	query.Reject(http.StatusNotImplemented, nil)
}

// deleteJobURI handles DELETE /{JobUri}
func (srv *AbstractServer) deleteJobURI(query *abstractServerQuery) {
	srv.finish(JobCanceled, JobCanceledByUser)
	query.WriteHeader(http.StatusOK)
}

// finish finishes the current job and updates server state
func (srv *AbstractServer) finish(state JobState, reason JobStateReason) {
	srv.lock.Lock()
	defer srv.lock.Unlock()

	srv.document.Close()
	srv.document = nil
	srv.status.State = ScannerIdle
	srv.status.Jobs[0].JobState = state
	if reason != UnknownJobStateReason {
		srv.status.Jobs[0].JobStateReasons = []JobStateReason{reason}
	}
}
