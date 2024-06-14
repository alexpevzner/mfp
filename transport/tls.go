// MFP       - Miulti-Function Printers and scanners toolkit
// TRANSPORT - Transport protocol implementation
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// TLS auto-detect

package transport

import (
	"net"
	"sync"
	"syscall"
)

// autoTLSListener wraps net.Listener and provides additional
// functionality by multiplexing incoming connections into
// plain (non-TLS) and encrypted (with TLS) connections.
//
// When created, two child listeners are returned. These child
// listeners receive plain/encrypted connections, respectively.
type autoTLSListener struct {
	lock             sync.Mutex            // Access lock
	wait             sync.Cond             // Wait queue
	haveAccepter     bool                  // Have accepting goroutine
	closed           bool                  // Listener is closed
	parent           net.Listener          // Parent listener
	plain, encrypted autoTLSListenerQueue  // Queues of connections
	pending          map[net.Conn]struct{} // Detect in progress
}

// autoTLSListenerChild is the child listener for autoTLSListener.
type autoTLSListenerChild struct {
	*autoTLSListener
	encrypted bool
}

// autoTLSListenerQueue is the queue of net.Conn connections.
type autoTLSListenerQueue struct {
	connections []net.Conn
}

// autoTLSWithSyscallConn represents net.Conn implementations
// that support SyscallConn() method.
type autoTLSWithSyscallConn interface {
	SyscallConn() (syscall.RawConn, error)
}

// NewAutoTLSListener accepts [net.Listener] on input and returns
// two net.Listeners. Incoming connections are automatically and
// transparently multiplexed between these two listeners.
//
// First listener received plain (non-TLS) connections, second
// receives encrypted connections.
//
// Multiplexing is based on prefetching first few bytes sent by
// a client and analyzing these bytes.
//
// Closing of any of returned listeners closes the parent listener
// and unblocks all goroutines waiting for incoming connections.
func NewAutoTLSListener(parent net.Listener) (plain, encrypted net.Listener) {
	atl := &autoTLSListener{
		parent:  parent,
		pending: make(map[net.Conn]struct{}),
	}

	atl.wait.L = &atl.lock

	plain = autoTLSListenerChild{atl, false}
	encrypted = autoTLSListenerChild{atl, true}

	return
}

// accept waits for a new connection.
//
// It receives all connections from the parent listener, classifies
// them as plain/encrypted and returns the connection of desired
// type as soon as it becomes available.
func (atl *autoTLSListener) accept(encrypted bool) (net.Conn, error) {
	// Choose queue we are interested in.
	queue := &atl.plain
	if encrypted {
		queue = &atl.encrypted
	}

	// Continue under lock.
	atl.lock.Lock()
	defer atl.lock.Unlock()

	for {
		// May be we already have a queued connection?
		c := queue.pull()
		if c != nil {
			return c, nil
		}

		// Somebody already waits on parent.Accept()?
		if atl.haveAccepter {
			atl.wait.Wait()
			continue
		}

		// We are that happy accepter.
		atl.haveAccepter = true
		atl.lock.Unlock()

		err := atl.acceptWait()

		atl.lock.Lock()
		atl.haveAccepter = false

		atl.wait.Broadcast()

		if err != nil {
			return nil, err
		}
	}
}

// close closes the listener and aborts all pending connections
func (atl *autoTLSListener) close() {
	atl.lock.Lock()

	atl.closed = true

	conns := atl.plain.pullAll()
	conns = append(conns, atl.encrypted.pullAll()...)
	for c := range atl.pending {
		delete(atl.pending, c)
		conns = append(conns, c)
	}

	atl.lock.Unlock()

	atl.parent.Close()

	for _, c := range conns {
		c.Close()
	}

}

// acceptWait waits for the next incoming connection on a parent listener.
// Then, on success, it calls connClassify() to push the connection into
// one of connections queue.
func (atl *autoTLSListener) acceptWait() error {
	c, err := atl.parent.Accept()
	if err == nil {
		// Add connection to atl.pending, so if listener will
		// be closed from another goroutine, read will unblock.
		atl.lock.Lock()

		closed := atl.closed
		if !closed {
			atl.pending[c] = struct{}{}
		}

		atl.lock.Unlock()

		// If listener already closed, just drop the connection
		if closed {
			c.Close()
			return nil
		}

		// Detect TLS, then drop connection from pending.
		withTLS, err := atl.detectTLS(c)

		atl.lock.Lock()
		delete(atl.pending, c)
		atl.lock.Unlock()

		// connClassify fails if read from connection fails.
		// It's not a problem of the entire listener, it is
		// a problem of this particular connection.
		//
		// So drop the connection and otherwise ignore an error.
		if err != nil {
			c.Close()
			return nil
		}

		// Enqueue the connection
		atl.lock.Lock()

		switch {
		case atl.closed:
			c.Close() // No more new connections
		case withTLS:
			atl.encrypted.push(c)
		default:
			atl.plain.push(c)
		}

		atl.lock.Unlock()

	}
	return err
}

// detectTLS detects if connection is encrypted or plain and.
//
// Detection requires few bytes of data to be fetched from the
// connection, and it may fail, so the function may return error.
func (atl *autoTLSListener) detectTLS(c net.Conn) (withTLS bool, err error) {
	conn, ok := c.(autoTLSWithSyscallConn)
	if ok {
		rawconn, err := conn.SyscallConn()
		if err == nil {
			return atl.detectTLSRawConn(rawconn)
		}
	}

	// FIXME - implement detectTLS on connections that
	// don't provide a SyscallConn() method.

	return false, nil
}

// detectTLSRawConn detects TLS on a syscall.RawConn.
func (atl *autoTLSListener) detectTLSRawConn(rawconn syscall.RawConn) (
	withTLS bool, err error) {

	buf := make([]byte, 16)

	rawconn.Read(func(fd uintptr) bool {
		var n int
		n, _, err = syscall.Recvfrom(int(fd), buf,
			syscall.MSG_PEEK)

		if n > 0 {
			buf = buf[:n]
			return true
		}

		if err != syscall.EAGAIN && err != syscall.EWOULDBLOCK {
			return true
		}

		return false
	})

	if err == nil {
		withTLS = buf[0] == 0x16
	}

	return withTLS, nil
}

// Accept waits for and returns the next connection to the listener.
func (l autoTLSListenerChild) Accept() (net.Conn, error) {
	return l.accept(l.encrypted)
}

// Close closes the listener.
func (l autoTLSListenerChild) Close() error {
	l.close()
	return nil
}

// Addr returns listener address.
func (l autoTLSListenerChild) Addr() net.Addr {
	return l.parent.Addr()
}

// push pushed connection to the queue.
func (q *autoTLSListenerQueue) push(c net.Conn) {
	q.connections = append(q.connections, c)
}

// pull returns next connection from the queue.
// If queue is empty, it returns nil.
func (q *autoTLSListenerQueue) pull() (c net.Conn) {
	if len(q.connections) > 0 {
		c = q.connections[0]
		copy(q.connections, q.connections[1:])
		q.connections = q.connections[:len(q.connections)-1]
	}
	return
}

// pullAll returns all the connections from the queue
// and purges the queue.
func (q *autoTLSListenerQueue) pullAll() []net.Conn {
	conns := q.connections
	q.connections = q.connections[:0]
	return conns
}
