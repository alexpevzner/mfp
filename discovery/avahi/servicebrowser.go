// MFP - Miulti-Function Printers and scanners toolkit
// Cgo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Service browser
//
//go:build linux || freebsd

package avahi

import (
	"runtime/cgo"
	"unsafe"
)

// #include <stdlib.h>
// #include <avahi-client/lookup.h>
//
// void serviceBrowserCallback (
//	AvahiServiceBrowser *b,
//	AvahiIfIndex interface,
//	AvahiProtocol proto,
//	AvahiBrowserEvent event,
//	const char *name,
//	const char *type,
//	const char *domain,
//	AvahiLookupResultFlags flags,
//	void *userdata);
import "C"

// ServiceBrowser returns available services of the specified type.
// Service type is a string that looks like "_http._tcp", "_ipp._tcp"
// and so on.
type ServiceBrowser struct {
	clnt         *Client                          // Owning Client
	handle       cgo.Handle                       // Handle to self
	avahiBrowser *C.AvahiServiceBrowser           // Underlying object
	queue        eventqueue[*ServiceBrowserEvent] // Event queue
}

// ServiceBrowserEvent represents events, generated by the
// [ServiceBrowser].
type ServiceBrowserEvent struct {
	Event    BrowserEvent      // Event code
	IfIndex  IfIndex           // Network interface index
	Protocol Protocol          // Network protocol
	Err      ErrCode           // In a case of BrowserFailure
	Flags    LookupResultFlags // Lookup flags
	Name     string            // Service name
	Type     string            // Service type
	Domain   string            // Service domain
}

// NewServiceBrowser creates a new [ServiceBrowser].
//
// ServiceBrowser constantly monitors the network and generates
// [ServiceBrowserEvent] events via channel returned by the
// [ServiceBrowser.Chan]
//
// Function parameters:
//   - clnt is the pointer to [Client]
//   - ifindex is the network interface index. Use [IfIndexUnspec]
//     to monitor all interfaces.
//   - proto is the IP4/IP6 protocol, used as transport for queries. If
//     set to [ProtocolUnspec], both protocols will be used.
//   - svctype is the service type we are looking for (e.g., "_http._tcp")
//   - domain is domain where service is looked. If set to "", the
//     default domain is used, which depends on a avahi-daemon configuration
//     and usually is ".local"
//   - flags provide some lookup options. See [LookupFlags] for details.
func NewServiceBrowser(
	clnt *Client,
	ifindex IfIndex,
	proto Protocol,
	svctype, domain string,
	flags LookupFlags) (*ServiceBrowser, error) {

	// Initialize ServiceBrowser structure
	browser := &ServiceBrowser{clnt: clnt}
	browser.handle = cgo.NewHandle(browser)
	browser.queue.init()

	// Convert strings from Go to C
	csvctype := C.CString(svctype)
	defer C.free(unsafe.Pointer(csvctype))

	var cdomain *C.char
	if domain != "" {
		cdomain = C.CString(domain)
		defer C.free(unsafe.Pointer(cdomain))
	}

	// Create AvahiServiceBrowser
	avahiClient := clnt.begin()
	defer clnt.end()

	browser.avahiBrowser = C.avahi_service_browser_new(
		avahiClient,
		C.AvahiIfIndex(ifindex),
		C.AvahiProtocol(proto),
		csvctype, cdomain,
		C.AvahiLookupFlags(flags),
		C.AvahiServiceBrowserCallback(C.serviceBrowserCallback),
		unsafe.Pointer(&browser.handle),
	)

	if browser.avahiBrowser == nil {
		browser.queue.Close()
		browser.handle.Delete()
		return nil, clnt.errno()
	}

	return browser, nil
}

// Chan returns channel where [ServiceBrowserEvent]s are sent.
func (browser *ServiceBrowser) Chan() <-chan *ServiceBrowserEvent {
	return browser.queue.Chan()
}

// Close closes the [ServiceBrowser] and releases allocated resources.
// It closes the event channel, effectively unblocking pending readers.
func (browser *ServiceBrowser) Close() {
	browser.clnt.begin()
	C.avahi_service_browser_free(browser.avahiBrowser)
	browser.avahiBrowser = nil
	browser.clnt.end()

	browser.queue.Close()
	browser.handle.Delete()
}

// serviceBrowserCallback called by AvahiServiceBrowser to
// report discovered services
func serviceBrowserCallback(
	b *C.AvahiServiceBrowser,
	ifindex C.AvahiIfIndex,
	proto C.AvahiProtocol,
	event BrowserEvent,
	name, svctype, domain *C.char,
	flags C.AvahiLookupResultFlags,
	p unsafe.Pointer) {

	browser := *(*cgo.Handle)(p).Value().(*ServiceBrowser)

	evnt := &ServiceBrowserEvent{
		Event:    BrowserEvent(event),
		IfIndex:  IfIndex(ifindex),
		Protocol: Protocol(proto),
		Err:      browser.clnt.errno(),
		Flags:    LookupResultFlags(flags),
		Name:     C.GoString(name),
		Type:     C.GoString(svctype),
		Domain:   C.GoString(domain),
	}

	browser.queue.Push(evnt)
}
