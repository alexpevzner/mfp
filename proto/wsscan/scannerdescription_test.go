// MFP - Multi-Function Printers and scanners toolkit
// WS-Scan core protocol
//
// Copyright (C) 2024 and up by Yogesh Singla (yogeshsingla481@gmail.com)
// See LICENSE for license terms and conditions
//
// Test for scanner description

package wsscan

import (
	"reflect"
	"testing"

	"github.com/OpenPrinting/go-mfp/internal/assert"
	"github.com/OpenPrinting/go-mfp/util/optional"
)

func TestScannerDescription(t *testing.T) {
	// Test with all child elements
	sd := ScannerDescription{
		ScannerName: ScannerName{
			Name: "Accounting Scanner in Copy Room 2",
			Lang: optional.New("en-AU, en-CA, en-GB, en-US"),
		},
		ScannerInfo: optional.New(ScannerInfo{
			Info: "High-speed document scanner for accounting department",
			Lang: optional.New("en-US"),
		}),
		ScannerLocation: optional.New(ScannerLocation{
			Location: "LA Campus - Building 1",
			Lang:     optional.New("en-AU, en-CA, en-GB, en-US"),
		}),
	}

	elm := sd.toXML(NsWSCN + ":ScannerDescription")
	if elm.Name != NsWSCN+":ScannerDescription" {
		t.Errorf("expected element name '%s:ScannerDescription', got '%s'", NsWSCN, elm.Name)
	}

	// Check that all child elements are present
	if len(elm.Children) != 3 {
		t.Errorf("expected 3 child elements, got %d", len(elm.Children))
	}

	// Test round-trip
	sd2, err := decodeScannerDescription(elm)
	assert.NoError(err)
	if !reflect.DeepEqual(sd, sd2) {
		t.Errorf("expected %v, got %v", sd, sd2)
	}

	// Test with only required ScannerName
	sdMinimal := ScannerDescription{
		ScannerName: ScannerName{
			Name: "Basic Scanner",
		},
	}

	elm2 := sdMinimal.toXML(NsWSCN + ":ScannerDescription")
	if len(elm2.Children) != 1 {
		t.Errorf("expected 1 child element, got %d", len(elm2.Children))
	}

	sd3, err := decodeScannerDescription(elm2)
	assert.NoError(err)
	if !reflect.DeepEqual(sdMinimal, sd3) {
		t.Errorf("expected %v, got %v", sdMinimal, sd3)
	}
}
