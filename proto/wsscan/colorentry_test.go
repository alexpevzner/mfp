// MFP - Multi-Function Printers and scanners toolkit
// wsscan core protocol
//
// Copyright (C) 2024 and up by Yogesh Singla (yogeshsingla481@gmail.com)
// See LICENSE for license terms and conditions
//
// Test for scan color entry

package wsscan

import "testing"

var testColorEntry = testEnum[ColorEntry]{
	decodeStr: DecodeColorEntry,
	decodeXML: decodeColorEntry,
	dataset: []testEnumData[ColorEntry]{
		{BlackAndWhite1, "BlackAndWhite1"},
		{Grayscale4, "Grayscale4"},
		{Grayscale8, "Grayscale8"},
		{Grayscale16, "Grayscale16"},
		{RGB24, "RGB24"},
		{RGB48, "RGB48"},
		{RGBa32, "RGBa32"},
		{RGBa64, "RGBa64"},
	},
}

// TestColorEntry tests [ColorEntry] common methods and functions.
func TestColorEntry(t *testing.T) {
	testColorEntry.run(t)
}
