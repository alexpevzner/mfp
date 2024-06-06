// MFP       - Miulti-Function Printers and scanners toolkit
// TRANSPORT - Transport protocol implementation
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Tests for IPP-specific URL parsing

package transport

import (
	"errors"
	"testing"
)

// TestParseUrl tests ParseURL function
func TestParseURL(t *testing.T) {
	type testData struct {
		in  string
		out string
		err string
	}

	tests := []testData{
		// HTTP schemes
		{
			in:  "http://127.0.0.1/ipp/print",
			out: "http://127.0.0.1/ipp/print",
		},

		{
			in:  "http://127.0.0.1:80/ipp/print",
			out: "http://127.0.0.1/ipp/print",
		},

		{
			in:  "http://127.0.0.1:81/ipp/print",
			out: "http://127.0.0.1:81/ipp/print",
		},

		{
			in:  "https://127.0.0.1:443/ipp/print",
			out: "https://127.0.0.1/ipp/print",
		},

		{
			in:  "https://127.0.0.1:444/ipp/print",
			out: "https://127.0.0.1:444/ipp/print",
		},

		{
			in:  "http://[fe80::aec5:1bff:fe1c:6fa7%252]/ipp/print",
			out: "http://[fe80::aec5:1bff:fe1c:6fa7%252]/ipp/print",
		},

		// IPP schemes
		{
			in:  "ipp://127.0.0.1/ipp/print",
			out: "ipp://127.0.0.1/ipp/print",
		},

		{
			in:  "ipp://127.0.0.1:631/ipp/print",
			out: "ipp://127.0.0.1/ipp/print",
		},

		{
			in:  "ipps://127.0.0.1:631/ipp/print",
			out: "ipps://127.0.0.1/ipp/print",
		},

		{
			in:  "ipps://127.0.0.1:632/ipp/print",
			out: "ipps://127.0.0.1:632/ipp/print",
		},

		// UNIX schema
		{
			in:  "unix:///var/run/cups/cups.sock",
			out: "unix:/var/run/cups/cups.sock",
		},

		{
			in:  "unix:/var/run/cups/cups.sock",
			out: "unix:/var/run/cups/cups.sock",
		},

		{
			in:  "unix://localhost/var/run/cups/cups.sock",
			out: "unix:/var/run/cups/cups.sock",
		},

		{
			in:  "unix://LoCaLhOsT/var/run/cups/cups.sock",
			out: "unix:/var/run/cups/cups.sock",
		},

		{
			in:  "unix://localhost:80/var/run/cups/cups.sock",
			err: ErrURLUNIXHost.Error(),
		},

		{
			in:  "unix://example.com/var/run/cups/cups.sock",
			err: ErrURLUNIXHost.Error(),
		},

		// Path handling
		{
			in:  "http://127.0.0.1/",
			out: "http://127.0.0.1/",
		},

		{
			in:  "http://127.0.0.1",
			out: "http://127.0.0.1/",
		},

		{
			in:  "http://127.0.0.1/foo/",
			out: "http://127.0.0.1/foo/",
		},

		{
			in:  "http://127.0.0.1/foo//////bar",
			out: "http://127.0.0.1/foo/bar",
		},

		{
			in:  "http://127.0.0.1/foo/./bar/../foobar",
			out: "http://127.0.0.1/foo/foobar",
		},

		// Scheme errors
		{
			in:  "foo",
			err: ErrURLSchemeMissed.Error(),
		},

		{
			in:  "foo:",
			err: ErrURLSchemeInvalid.Error(),
		},

		// Other errors:
		{
			in:  "http://Invalid URL",
			err: ErrURLInvalid.Error(),
		},

		{
			in:  "",
			err: ErrURLSchemeMissed.Error(),
		},

		{
			in:  "unix:///var/run/cups/cups.sock",
			out: "unix:/var/run/cups/cups.sock",
		},

		{
			in:  "unix:/var/run/cups/cups.sock",
			out: "unix:/var/run/cups/cups.sock",
		},
	}

	for _, test := range tests {
		u, err := ParseURL(test.in)
		if err == nil {
			err = errors.New("")
		}

		switch {
		case err.Error() != test.err:
			t.Errorf("%q: error mismatch:\nexpected: %s\npresent:  %s",
				test.in, test.err, err)

		case test.err != "":
			// Error as expected; nothing to do

		case u.String() != test.out:
			t.Errorf("%q: output mismatch:\nexpected: %s\npresent:  %s",
				test.in, test.out, u)
		}
	}
}

// TestMustParseURL tests how MustParseURL panics in a case of errors
func TestMustParseURL(t *testing.T) {
	defer func() {
		err := recover()
		if err != ErrURLSchemeMissed {
			t.Errorf("Error expected: %q, present: %v",
				ErrURLSchemeMissed, err)
		}
	}()

	MustParseURL("foo")
}

// TestParseAddr tests ParseAddr function
func TestParseAddr(t *testing.T) {
	type testData struct {
		in       string // Input address
		template string // Template URL
		out      string // Expected output
		err      string // Expected error
	}

	tests := []testData{
		// IP4 and IP4 addresses
		{
			in:  "127.0.0.1",
			out: "http://127.0.0.1/",
		},

		{
			in:  "::1",
			out: "http://[::1]/",
		},

		{
			in:  "[::1]",
			out: "http://[::1]/",
		},

		// IP4 and IP4 addresses with port
		{
			in:  "127.0.0.1:80",
			out: "http://127.0.0.1/",
		},

		{
			in:  "127.0.0.1:81",
			out: "http://127.0.0.1:81/",
		},

		{
			in:  "127.0.0.1:443",
			out: "https://127.0.0.1/",
		},

		{
			in:  "127.0.0.1:631",
			out: "ipp://127.0.0.1/",
		},

		{
			in:  "[::1]:80",
			out: "http://[::1]/",
		},

		{
			in:  "[::1]:81",
			out: "http://[::1]:81/",
		},

		// UNIX paths
		{
			in:  "/var/run/cups/cups.sock",
			out: "unix:/var/run/cups/cups.sock",
		},

		// IP address with template
		{
			in:       "127.0.0.1",
			template: "https://localhost/",
			out:      "https://127.0.0.1/",
		},

		{
			in:       "127.0.0.1",
			template: "http://localhost:222/",
			out:      "http://127.0.0.1:222/",
		},

		// IP address and port with template
		{
			in:       "127.0.0.1:1234",
			template: "https://localhost/path",
			out:      "https://127.0.0.1:1234/path",
		},

		// Full URLs
		{
			in:  "http://127.0.0.1/ipp/print",
			out: "http://127.0.0.1/ipp/print",
		},

		{
			in:  "http://127.0.0.1:80/ipp/print",
			out: "http://127.0.0.1/ipp/print",
		},
	}

	for _, test := range tests {
		u, err := ParseAddr(test.in, test.template)
		if err == nil {
			err = errors.New("")
		}

		switch {
		case err.Error() != test.err:
			t.Errorf("%q: error mismatch:\nexpected: %s\npresent:  %s",
				test.in, test.err, err)

		case test.err != "":
			// Error as expected; nothing to do

		case u.String() != test.out:
			t.Errorf("%q: output mismatch:\nexpected: %s\npresent:  %s",
				test.in, test.out, u)
		}

	}
}
