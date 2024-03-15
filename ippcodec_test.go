// IPPX - High-level implementation of IPP printing protocol on Go
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// IPP codec test

package ippx

import (
	"errors"
	"reflect"
	"testing"

	"github.com/OpenPrinting/goipp"
)

// ippEncodeDecodeTest represents a single IPP encode/decode test
type ippEncodeDecodeTest struct {
	t        reflect.Type // Input type
	data     interface{}  // Input data
	panic    error        // Expected panic
	encError error        // Expected encode error
}

// ippEncodeDecodeTestData is the test data for the IPP encode/decode test
var ippEncodeDecodeTestData = []ippEncodeDecodeTest{
	{
		t:     reflect.TypeOf(int(0)),
		data:  1,
		panic: errors.New(`int is not struct`),
	},
	{
		t:     reflect.TypeOf(PrinterAttributes{}),
		data:  1,
		panic: errors.New(`Encoder for "*PrinterAttributes" applied to "int"`),
	},
	{
		t:    reflect.TypeOf(PrinterAttributes{}),
		data: &testdataPrinterAttributes,
	},
}

func (test ippEncodeDecodeTest) exec(t *testing.T) {
	// This function catches the possible panic
	defer func() {
		// Panic not expected - let it go its way
		if test.panic == nil {
			return
		}

		p := recover()
		if p == nil && test.panic != nil {
			t.Errorf("panic expected but didn't happen: %s",
				test.panic)
			return
		}

		if p != nil {
			err, ok := p.(error)
			if !ok {
				panic(p)
			}

			if err.Error() != test.panic.Error() {
				t.Errorf("panic expected: %s, got: %s",
					test.panic, err)
			}
		}
	}()

	// Generate codec
	codec := ippCodecMustGenerate(test.t)

	// Test encoding
	var attrs goipp.Attributes
	err := codec.encode(test.data, &attrs)

	checkError(t, err, test.encError)
	if err != nil {
		return
	}

	// Test decoding
	out := reflect.New(test.t).Interface()
	err = codec.decode(out, attrs)

	checkError(t, err, test.encError)
	if err != nil {
		return
	}

	if !reflect.DeepEqual(test.data, out) {
		t.Errorf("input/output mismatch")
	}
}

// IPP encode/decode test
func TestIppEncodeDecode(t *testing.T) {
	for _, test := range ippEncodeDecodeTestData {
		test.exec(t)
	}
}

// Check error against expected
func checkError(t *testing.T, err, expected error) {
	switch {
	case err == nil && expected != nil:
		t.Errorf("error expected but didn't happen: %s", expected)
	case err != nil && expected == nil:
		t.Errorf("error not expected: %s", err)
	case err != nil && expected != nil && err.Error() != expected.Error():
		t.Errorf("error expected: %s, got: %s", expected, err)
	}
}

var testdataPrinterAttributes = PrinterAttributes{
	CharsetConfigured:    DefaultCharsetConfigured,
	CharsetSupported:     DefaultCharsetSupported,
	CompressionSupported: []string{"none"},
	IppFeaturesSupported: []string{
		"airprint-1.7",
		"airprint-1.6",
		"airprint-1.5",
		"airprint-1.4",
	},
	IppVersionsSupported: DefaultIppVersionsSupported,
	MediaSizeSupported: []PrinterMediaSizeSupported{
		{21590, 27940},
		{21000, 29700},
	},
	MediaSizeSupportedRange: PrinterMediaSizeSupportedRange{
		XDimension: goipp.Range{Lower: 10000, Upper: 14800},
		YDimension: goipp.Range{Lower: 21600, Upper: 35600},
	},
	OperationsSupported: []goipp.Op{
		goipp.OpGetPrinterAttributes,
	},
}
