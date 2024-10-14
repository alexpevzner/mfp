// MFP - Miulti-Function Printers and scanners toolkit
// WSD device discovery
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// WSDD backend

package wsdd

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/alexpevzner/mfp/discovery"
	"github.com/alexpevzner/mfp/discovery/netstate"
	"github.com/alexpevzner/mfp/log"
	"github.com/alexpevzner/mfp/wsd"
)

// backend is the [discovery.Backend] for WSD device discovery.
type backend struct {
	ctx     context.Context       // For logging and backend.Close
	cancel  context.CancelFunc    // Context's cancel function
	queue   *discovery.Eventqueue // Event queue
	netmon  *netstate.Notifier    // Network state monitor
	mconn4  *mconn                // IP4 multicasts reception connection
	mconn6  *mconn                // IP6 multicasts reception connection
	closing atomic.Bool           // Close in progress
	done    sync.WaitGroup        // For backend.Close synchronization
}

// NewBackend creates a new [discovery.Backend] for WSD device discovery.
func NewBackend(ctx context.Context) (discovery.Backend, error) {
	// Set log prefix
	ctx = log.WithPrefix(ctx, "wsdd")

	// Create multicast sockets
	mconn4, err := newMconn(wsddMulticastIP4)
	if err != nil {
		return nil, err
	}

	mconn6, err := newMconn(wsddMulticastIP6)
	if err != nil {
		mconn4.Close()
		return nil, err
	}

	// Create cancelable context
	ctx, cancel := context.WithCancel(ctx)

	// Create backend structure
	back := &backend{
		ctx:    ctx,
		cancel: cancel,
		netmon: netstate.NewNotifier(),
		mconn4: mconn4,
		mconn6: mconn6,
	}
	return back, nil
}

// Name returns backend name.
func (back *backend) Name() string {
	return "wsdd"
}

// Start starts Backend operations.
func (back *backend) Start(queue *discovery.Eventqueue) {
	back.queue = queue

	back.done.Add(3)

	go back.netmonProc()
	go back.mconnProc(back.mconn4)
	go back.mconnProc(back.mconn6)

	log.Debug(back.ctx, "backend started")
}

// Close closes the backend
func (back *backend) Close() {
	back.closing.Store(true)
	back.cancel()
	back.mconn4.Close()
	back.mconn6.Close()
	back.done.Wait()
}

// netmonproc processes netstate.Notifier events.
// It runs on its own goroutine.
func (back *backend) netmonProc() {
	defer back.done.Done()

	for {
		evnt, err := back.netmon.Get(back.ctx)
		if err != nil {
			return
		}

		log.Debug(back.ctx, "%s", evnt)
	}
}

// mconn4proc receives UDP multicast messages from the multicast conection.
// the back.mconn4 connection.
func (back *backend) mconnProc(mc *mconn) {
	defer back.done.Done()

	for {
		var buf [65536]byte
		n, from, cmsg, err := mc.RecvFrom(buf[:])

		if back.closing.Load() {
			return
		}

		if err != nil {
			log.Error(back.ctx, "UDP recv: %s", err)
			return
		}

		log.Debug(back.ctx, "%d bytes received from %s%%%d",
			n, from, cmsg.IfIndex)

		data := buf[:n]
		msg, err := wsd.DecodeMsg(data)
		if err != nil {
			log.Warning(back.ctx, "%s", err)
			continue
		}

		log.Debug(back.ctx, "%s message received", msg.Header.Action)
	}
}
