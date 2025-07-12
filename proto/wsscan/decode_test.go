// MFP - Miulti-Function Printers and scanners toolkit
// WS-Scan core protocol
//
// Copyright (C) 2024 and up by Yogesh Singla (yogeshsingla481@gmail.com)
// See LICENSE for license terms and conditions
//
// Tests for the basic XML decoding functions

package wsscan

import (
	"errors"
	"testing"

	"github.com/OpenPrinting/go-mfp/util/xmldoc"
)

// TestDecodeInt tests decodeInt
func TestDecodeInt(t *testing.T) {
	type testData struct {
		xml xmldoc.Element // Input data
		out int            // Expected output
		err string         // Expected error
	}

	tests := []testData{
		{
			xml: xmldoc.WithText("test", "123"),
			out: 123,
			err: ``,
		},

		{
			xml: xmldoc.WithText("test", "0"),
			out: 0,
			err: ``,
		},

		{
			xml: xmldoc.WithText("test", "2147483647"),
			out: 2147483647,
			err: ``,
		},

		{
			xml: xmldoc.WithText("test", "-1"),
			out: -1,
			err: ``,
		},

		{
			xml: xmldoc.WithText("test", "2147483648"),
			err: `/test: int out of range: 2147483648`,
		},

		{
			xml: xmldoc.WithText("test", "bad"),
			err: `/test: invalid int: "bad"`,
		},
	}

	for _, test := range tests {
		out, err := decodeInt(test.xml)
		if err == nil {
			err = errors.New("")
		}

		if err.Error() != test.err {
			t.Errorf("error mismatch:\n"+
				"input:          %s\n"+
				"error expected: %q\n"+
				"error present:  %q\n",
				test.xml.EncodeString(nil), test.err, err)
		}

		if err.Error() == "" && out != test.out {
			t.Errorf("output mismatch:\n"+
				"input:    %s\n"+
				"expected: %v\n"+
				"present:  %v\n",
				test.xml.EncodeString(nil), test.out, out)
		}
	}
}

// TestDecodeNonNegativeInt tests decodeNonNegativeInt
func TestDecodeNonNegativeInt(t *testing.T) {
	type testData struct {
		xml xmldoc.Element // Input data
		out int            // Expected output
		err string         // Expected error
	}

	tests := []testData{
		{
			xml: xmldoc.WithText("test", "123"),
			out: 123,
			err: ``,
		},

		{
			xml: xmldoc.WithText("test", "0"),
			out: 0,
			err: ``,
		},

		{
			xml: xmldoc.WithText("test", "2147483647"),
			out: 2147483647,
			err: ``,
		},

		{
			xml: xmldoc.WithText("test", "-1"),
			err: `/test: int out of range: -1`,
		},

		{
			xml: xmldoc.WithText("test", "2147483648"),
			err: `/test: int out of range: 2147483648`,
		},

		{
			xml: xmldoc.WithText("test", "bad"),
			err: `/test: invalid int: "bad"`,
		},
	}

	for _, test := range tests {
		out, err := decodeNonNegativeInt(test.xml)
		if err == nil {
			err = errors.New("")
		}

		if err.Error() != test.err {
			t.Errorf("error mismatch:\n"+
				"input:          %s\n"+
				"error expected: %q\n"+
				"error present:  %q\n",
				test.xml.EncodeString(nil), test.err, err)
		}

		if err.Error() == "" && out != test.out {
			t.Errorf("output mismatch:\n"+
				"input:    %s\n"+
				"expected: %v\n"+
				"present:  %v\n",
				test.xml.EncodeString(nil), test.out, out)
		}
	}
}

// TestDecodeBool tests decodeBool
func TestDecodeBool(t *testing.T) {
	type testData struct {
		xml xmldoc.Element // Input data
		out bool           // Expected output
		err string         // Expected error
	}

	tests := []testData{
		{
			xml: xmldoc.WithText("test", "true"),
			out: true,
			err: ``,
		},

		{
			xml: xmldoc.WithText("test", "false"),
			out: false,
			err: ``,
		},

		{
			xml: xmldoc.WithText("test", "bad"),
			err: `/test: invalid bool: "bad"`,
		},
	}

	for _, test := range tests {
		out, err := decodeBool(test.xml)
		if err == nil {
			err = errors.New("")
		}

		if err.Error() != test.err {
			t.Errorf("error mismatch:\n"+
				"input:          %s\n"+
				"error expected: %q\n"+
				"error present:  %q\n",
				test.xml.EncodeString(nil), test.err, err)
		}

		if err.Error() == "" && out != test.out {
			t.Errorf("output mismatch:\n"+
				"input:    %s\n"+
				"expected: %v\n"+
				"present:  %v\n",
				test.xml.EncodeString(nil), test.out, out)
		}
	}
}
