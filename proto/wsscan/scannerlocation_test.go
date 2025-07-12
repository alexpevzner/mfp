// MFP - Multi-Function Printers and scanners toolkit
// WS-Scan core protocol
//
// Copyright (C) 2024 and up by Yogesh Singla (yogeshsingla481@gmail.com)
// See LICENSE for license terms and conditions
//
// Test for scanner location

package wsscan

import (
	"reflect"
	"testing"

	"github.com/OpenPrinting/go-mfp/internal/assert"
	"github.com/OpenPrinting/go-mfp/util/optional"
)

func TestScannerLocation(t *testing.T) {
	dataset := []ScannerLocation{
		{Location: "LA Campus - Building 1", Lang: optional.New("en-AU, en-CA, en-GB, en-US")},
		{Location: "Office Floor 3", Lang: optional.New("en-US")},
		{Location: "Reception Area"},
	}

	for _, sl := range dataset {
		elm := sl.toXML("wscn:ScannerLocation")
		if elm.Name != "wscn:ScannerLocation" {
			t.Errorf("expected element name 'wscn:ScannerLocation', got '%s'", elm.Name)
		}
		if elm.Text != sl.Location {
			t.Errorf("expected element text '%s', got '%s'", sl.Location, elm.Text)
		}

		sl2, err := decodeScannerLocation(elm)
		assert.NoError(err)
		if !reflect.DeepEqual(sl, sl2) {
			t.Errorf("expected %v, got %v", sl, sl2)
		}
	}
}
