// MFP - Miulti-Function Printers and scanners toolkit
// Network state monitoring
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Test of set of network interfaces

package netstate

import (
	"fmt"
	"testing"
)

// TestSetOfInterfaces tests setOfInterfaces
func TestSetOfInterfaces(t *testing.T) {
	type testData struct {
		nif NetIf  // Interface to add/del
		op  string // Operation ("add", "del" or "contains")
		ret int    // Expected return value
	}

	netifmaker := testNewNetIfMaker()
	if0 := netifmaker.new()
	if1 := netifmaker.new()

	set := newSetOfInterfaces()

	tests := []testData{
		{
			nif: if0,
			op:  "add",
			ret: 0,
		},

		{
			nif: if0,
			op:  "contains",
			ret: 1,
		},

		{
			nif: if1,
			op:  "add",
			ret: 0,
		},

		{
			nif: if1,
			op:  "del",
			ret: 1,
		},

		{
			nif: if1,
			op:  "contains",
			ret: 0,
		},

		{
			nif: if0,
			op:  "add",
			ret: 1,
		},

		{
			nif: if0,
			op:  "add",
			ret: 2,
		},

		{
			nif: if0,
			op:  "contains",
			ret: 3,
		},
	}

	for i, test := range tests {
		nif := test.nif
		var ret int
		switch test.op {
		case "add":
			ret = set.add(nif)
		case "del":
			ret = set.del(nif)
		case "contains":
			ret = set.contains(nif)
		default:
			panic(fmt.Errorf("invalid operation %q", test.op))
		}

		if ret != test.ret {
			t.Errorf("%d: %s %s:\n"+
				"expected: %d\n"+
				" present: %d",
				i, test.op, nif.Name(), test.ret, ret)

		}
	}
}

// TestSetOfInterfacesAddAddrs tests setOfInterfaces.addAddrs
func TestSetOfInterfacesAddAddrs(t *testing.T) {
	netifmaker := testNewNetIfMaker()
	if0 := netifmaker.new()
	if1 := netifmaker.new()

	addrs := []Addr{
		testMakeAddr(if0, "192.168.0.1/24"),
		testMakeAddr(if1, "192.168.1.1/24"),
		testMakeAddr(if0, "192.168.0.2/24"),
		testMakeAddr(if1, "192.168.1.2/24"),
		testMakeAddr(if0, "192.168.0.3/24"),
	}

	set := newSetOfInterfaces()
	set.addAddrs(addrs)

	cnt0 := set.contains(if0)
	cnt1 := set.contains(if1)

	if cnt0 != 3 {
		t.Errorf("%s: expected 3, present %d", if0.Name(), cnt0)
	}

	if cnt1 != 2 {
		t.Errorf("%s: expected 3, present %d", if0.Name(), cnt0)
	}
}
