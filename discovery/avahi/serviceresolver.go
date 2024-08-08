// MFP - Miulti-Function Printers and scanners toolkit
// Cgo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Service resolver
//
//go:build linux || freebsd

package avahi

import (
	"context"
	"net/netip"
	"runtime/cgo"
	"unsafe"
)

// #include <stdlib.h>
// #include <avahi-client/lookup.h>
//
// void serviceResolverCallback (
//	AvahiServiceResolver *r,
//	AvahiIfIndex interface,
//	AvahiProtocol proto,
//	AvahiResolverEvent event,
//	char *name,
//	char *type,
//	char *domain,
//	char *host_name,
//	AvahiAddress *a,
//	uint16_t port,
//	AvahiStringList *txt,
//	AvahiLookupResultFlags flags,
//	void *userdata);
import "C"

// ServiceResolver resolves hostname, IP address and TXT record of
// the discovered services.
type ServiceResolver struct {
	clnt          *Client                           // Owning Client
	handle        cgo.Handle                        // Handle to self
	avahiResolver *C.AvahiServiceResolver           // Underlying object
	queue         eventqueue[*ServiceResolverEvent] // Event queue
}

// ServiceResolverEvent represents events, generated by the
// [ServiceResolver].
type ServiceResolverEvent struct {
	Event    ResolverEvent     // Event code
	IfIndex  IfIndex           // Network interface index
	Protocol Protocol          // Network protocol
	Err      ErrCode           // In a case of ResolverFailure
	Flags    LookupResultFlags // Lookup flags
	Name     string            // Service name
	Type     string            // Service type
	Domain   string            // Service domain
	Hostname string            // Resolved hostname
	Addr     netip.AddrPort    // Resolved IP address:port
	Txt      []string          // TXT record ("key=value"...)
}

// NewServiceResolver creates a new [ServiceResolver].
//
// ServiceResolver resolves hostname, IP address and TXT record of
// the services, previously discovered by the [ServiceBrowser] by
// service instance name ([ServiceBrowserEvent.InstanceName]).
// Resolved information is reported via channel returned by the
// [ServiceResolver.Chan].
//
// If IP address and/or TXT record is not needed, resolving of these
// parameters may be suppressed, using LookupNoAddress/LookupNoTXT
// [LookupFlags].
//
// Please notice, it is a common practice to register a service
// with a zero port value as a "placeholder" for missed service.
// For example, printers always register the "_printer._tcp" service
// to reserve the service name, but if LPD protocol is actually not
// supported, it will be registered with zero port.
//
// This is important to understand the proper usage of the "proto"
// and "addrproto" parameters and difference between them.  Please
// read the "IP4 vs IP6" section of the package Overview for technical
// details.
//
// Function parameters:
//   - clnt is the pointer to [Client]
//   - ifindex is the network interface index. Use [IfIndexUnspec]
//     to specify all interfaces.
//   - proto is the IP4/IP6 protocol, used as transport for queries. If
//     set to [ProtocolUnspec], both protocols will be used.
//   - instname is the service instance name, as reported by
//     [ServiceBrowserEvent.InstanceName]
//   - svctype is the service type we are looking for (e.g., "_http._tcp")
//   - domain is domain where service is looked. If set to "", the
//     default domain is used, which depends on a avahi-daemon configuration
//     and usually is ".local"
//   - addrproto specifies a protocol family of IP addresses we are
//     interested in. See explanation above for details.
//   - flags provide some lookup options. See [LookupFlags] for details.
//
// ServiceResolver must be closed after use with the [ServiceResolver.Close]
// function call.
func NewServiceResolver(
	clnt *Client,
	ifindex IfIndex,
	proto Protocol,
	instname, svctype, domain string,
	addrproto Protocol,
	flags LookupFlags) (*ServiceResolver, error) {

	// Initialize ServiceResolver structure
	resolver := &ServiceResolver{clnt: clnt}
	resolver.handle = cgo.NewHandle(resolver)
	resolver.queue.init()

	// Convert strings from Go to C
	cinstname := C.CString(instname)
	defer C.free(unsafe.Pointer(cinstname))

	csvctype := C.CString(svctype)
	defer C.free(unsafe.Pointer(csvctype))

	cdomain := C.CString(domain)
	defer C.free(unsafe.Pointer(cdomain))

	// Create AvahiServiceResolver
	avahiClient := clnt.begin()
	defer clnt.end()

	resolver.avahiResolver = C.avahi_service_resolver_new(
		avahiClient,
		C.AvahiIfIndex(ifindex),
		C.AvahiProtocol(proto),
		cinstname, csvctype, cdomain,
		C.AvahiProtocol(addrproto),
		C.AvahiLookupFlags(flags),
		C.AvahiServiceResolverCallback(C.serviceResolverCallback),
		unsafe.Pointer(&resolver.handle),
	)

	if resolver.avahiResolver == nil {
		resolver.queue.Close()
		resolver.handle.Delete()
		return nil, clnt.errno()
	}

	return resolver, nil
}

// Chan returns channel where [ServiceResolverEvent]s are sent.
func (resolver *ServiceResolver) Chan() <-chan *ServiceResolverEvent {
	return resolver.queue.Chan()
}

// Get waits for the next [ServiceResolverEvent].
//
// It returns:
//   - event, nil - if event available
//   - nil, error - if context is canceled
//   - nil, nil   - if ServiceResolver was closed
func (resolver *ServiceResolver) Get(ctx context.Context) (
	*ServiceResolverEvent, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case evnt := <-resolver.Chan():
		return evnt, nil
	}
}

// Close closes the [ServiceResolver] and releases allocated resources.
// It closes the event channel, effectively unblocking pending readers.
func (resolver *ServiceResolver) Close() {
	resolver.clnt.begin()
	C.avahi_service_resolver_free(resolver.avahiResolver)
	resolver.avahiResolver = nil
	resolver.clnt.end()

	resolver.queue.Close()
	resolver.handle.Delete()
}

// serviceResolverCallback called by AvahiServiceResolver to
// report discovered services
//
//export serviceResolverCallback
func serviceResolverCallback(
	r *C.AvahiServiceResolver,
	ifindex C.AvahiIfIndex,
	proto C.AvahiProtocol,
	event C.AvahiResolverEvent,
	name, svctype, domain, hostname *C.char,
	caddr *C.AvahiAddress,
	cport C.uint16_t,
	ctxt *C.AvahiStringList,
	flags C.AvahiLookupResultFlags,
	p unsafe.Pointer) {

	resolver := (*cgo.Handle)(p).Value().(*ServiceResolver)

	// Decode IP address:port
	var ip netip.Addr
	if caddr.proto == C.AVAHI_PROTO_INET {
		ip = netip.AddrFrom4(*(*[4]byte)(unsafe.Pointer(&caddr.data)))
	} else {
		ip = netip.AddrFrom16(*(*[16]byte)(unsafe.Pointer(&caddr.data)))
	}
	addr := netip.AddrPortFrom(ip, uint16(cport))

	// Decode TXT record
	var txt []string
	for ctxt != nil {
		t := C.GoStringN((*C.char)(unsafe.Pointer(&ctxt.text)),
			C.int(ctxt.size))
		txt = append(txt, t)

		ctxt = ctxt.next
	}

	// Generate an event
	evnt := &ServiceResolverEvent{
		Event:    ResolverEvent(event),
		IfIndex:  IfIndex(ifindex),
		Protocol: Protocol(proto),
		Err:      resolver.clnt.errno(),
		Flags:    LookupResultFlags(flags),
		Name:     C.GoString(name),
		Type:     C.GoString(svctype),
		Domain:   C.GoString(domain),
		Hostname: C.GoString(hostname),
		Addr:     addr,
		Txt:      txt,
	}

	resolver.queue.Push(evnt)
}
