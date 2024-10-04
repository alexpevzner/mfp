// MFP - Miulti-Function Printers and scanners toolkit
// WSD device discovery
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// UDP multicasting

package wsdd

import (
	"fmt"
	"net"
	"net/netip"
	"syscall"

	"github.com/alexpevzner/mfp/discovery/netstate"
)

// mconn wraps net.UDPConn and prepares it to be used for
// the UDP multicasts reception.
type mconn struct {
	*net.UDPConn
	group netip.Addr
}

// newMconn creates a new multicast connection
func newMconn(group netip.AddrPort) (*mconn, error) {
	// Address must be multicast
	if !group.Addr().IsMulticast() {
		err := fmt.Errorf("%s not multicast", group.Addr())
		return nil, err
	}

	// Prepare net.UDPAddr structure
	addr := &net.UDPAddr{
		IP:   net.IP(group.Addr().AsSlice()),
		Port: int(group.Port()),
		Zone: group.Addr().Zone(),
	}

	// Open UDP connection.
	//
	// Note, with the multicast address being given,
	// net.ListenUDP creates UDP socket bound to the
	// 0.0.0.0:port (or [::0]:port) address with
	// SO_REUSEADDR option being set.
	//
	// This socket can be joined multiple multicast
	// groups and suitable for the multicast reception.
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}

	// Fill and return mconn structure
	mc := &mconn{
		UDPConn: conn,
		group:   group.Addr(),
	}

	return mc, nil
}

// Join joins the multicast group, specified during mcast
// creation, on a network interface, specified by the local
// parameter.
func (mc *mconn) Join(local netstate.Addr) error {
	if mc.group.Is6() {
		return mc.joinIP6(local)
	}
	return mc.joinIP4(local)
}

// Leave leaves the multicast group, specified during mcast
// creation, on a network interface, specified by the local
// parameter.
func (mc *mconn) Leave(local netstate.Addr) error {
	if mc.group.Is6() {
		return mc.leaveIP6(local)
	}
	return mc.leaveIP4(local)
}

// joinIP4 is the mcast.Join for IP4 connections
func (mc *mconn) joinIP4(local netstate.Addr) error {
	if !mc.group.Is4() {
		err := fmt.Errorf("Can't join IP4 group on IP6 connection")
		return err
	}

	mreq := syscall.IPMreqn{
		Multiaddr: mc.group.As4(),
		Address:   local.Addr().As4(),
		Ifindex:   int32(local.Interface().Index()),
	}

	err := mc.control(func(fd int) error {
		return syscall.SetsockoptIPMreqn(fd, syscall.IPPROTO_IP,
			syscall.IP_ADD_MEMBERSHIP, &mreq)
	})

	return err
}

// joinIP6 is the mcast.Join for IP6 connections
func (mc *mconn) joinIP6(local netstate.Addr) error {
	if !mc.group.Is6() {
		err := fmt.Errorf("Can't join IP4 group on IP6 connection")
		return err
	}

	mreq := syscall.IPv6Mreq{
		Multiaddr: mc.group.As16(),
		Interface: uint32(local.Interface().Index()),
	}

	err := mc.control(func(fd int) error {
		return syscall.SetsockoptIPv6Mreq(fd, syscall.IPPROTO_IPV6,
			syscall.IPV6_JOIN_GROUP, &mreq)
	})

	return err
}

// leaveIP4 is the mcast.Leave for IP4 connections
func (mc *mconn) leaveIP4(local netstate.Addr) error {
	if !mc.group.Is4() {
		err := fmt.Errorf("Can't leave IP4 group on IP6 connection")
		return err
	}

	mreq := syscall.IPMreqn{
		Multiaddr: mc.group.As4(),
		Address:   local.Addr().As4(),
		Ifindex:   int32(local.Interface().Index()),
	}

	err := mc.control(func(fd int) error {
		return syscall.SetsockoptIPMreqn(fd, syscall.IPPROTO_IP,
			syscall.IP_DROP_MEMBERSHIP, &mreq)
	})

	return err
}

// leaveIP6 is the mcast.Leave for IP6 connections
func (mc *mconn) leaveIP6(local netstate.Addr) error {
	if !mc.group.Is6() {
		err := fmt.Errorf("Can't leave IP4 group on IP6 connection")
		return err
	}

	mreq := syscall.IPv6Mreq{
		Multiaddr: mc.group.As16(),
		Interface: uint32(local.Interface().Index()),
	}

	err := mc.control(func(fd int) error {
		return syscall.SetsockoptIPv6Mreq(fd, syscall.IPPROTO_IPV6,
			syscall.IPV6_LEAVE_GROUP, &mreq)
	})

	return err
}

// control invokes f on the underlying connection's
// file descriptor.
func (mc *mconn) control(f func(fd int) error) error {
	rawconn, err := mc.SyscallConn()
	if err != nil {
		return err
	}

	var err2 error
	err = rawconn.Control(func(fd uintptr) {
		err2 = f(int(fd))
	})

	if err != nil {
		return err
	}

	return err2
}
