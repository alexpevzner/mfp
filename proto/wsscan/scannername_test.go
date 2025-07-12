// MFP - Multi-Function Printers and scanners toolkit
// WS-Scan core protocol
//
// Copyright (C) 2024 and up by Yogesh Singla (yogeshsingla481@gmail.com)
// See LICENSE for license terms and conditions
//
// Test for scanner name

package wsscan

import (
	"reflect"
	"testing"

	"github.com/OpenPrinting/go-mfp/internal/assert"
	"github.com/OpenPrinting/go-mfp/util/optional"
)

func TestScannerName(t *testing.T) {
	dataset := []ScannerName{
		{Name: "Accounting Scanner in Copy Room 2", Lang: optional.New("en-AU, en-CA, en-GB, en-US")},
		{Name: "Main Office Scanner", Lang: optional.New("en-US")},
		{Name: "Reception Scanner"},
	}

	for _, sn := range dataset {
		elm := sn.toXML("wscn:ScannerName")
		if elm.Name != "wscn:ScannerName" {
			t.Errorf("expected element name 'wscn:ScannerName', got '%s'", elm.Name)
		}
		if elm.Text != sn.Name {
			t.Errorf("expected element text '%s', got '%s'", sn.Name, elm.Text)
		}

		sn2, err := decodeScannerName(elm)
		assert.NoError(err)
		if !reflect.DeepEqual(sn, sn2) {
			t.Errorf("expected %v, got %v", sn, sn2)
		}
	}
}
