// MFP - Miulti-Function Printers and scanners toolkit
// wsscan core protocol
//
// Copyright (C) 2024 and up by Yogesh Singla (yogeshsingla481@gmail.com)
// See LICENSE for license terms and conditions
//
// scan color entry

package wsscan

import (
	"github.com/OpenPrinting/go-mfp/util/xmldoc"
)

// ColorEntry defines the color mode of a scanned image, e.g., 1-bit B/W or 8-bit grayscale.
type ColorEntry int

// known color entries:
const (
	UnknownColorEntry ColorEntry = iota
	BlackAndWhite1               // 1 bpp, single channel
	Grayscale4                   // 4 bpp, single channel
	Grayscale8                   // 8 bpp, single channel
	Grayscale16                  // 16 bpp, single channel
	RGB24                        // 24 bpp, 3 channels (8 bits each)
	RGB48                        // 48 bpp, 3 channels (16 bits each)
	RGBa32                       // 32 bpp, 4 channels (8 bits each)
	RGBa64                       // 64 bpp, 4 channels (16 bits each)
)

// decodeColorEntry decodes [ColorEntry] from the XML tree.
func decodeColorEntry(root xmldoc.Element) (ce ColorEntry, err error) {
	return decodeEnum(root, DecodeColorEntry)
}

// toXML generates XML tree for the [ColorEntry].
func (ce ColorEntry) toXML(name string) xmldoc.Element {
	return xmldoc.Element{
		Name: name,
		Text: ce.String(),
	}
}

// String returns a string representation of the [ColorEntry]
func (ce ColorEntry) String() string {
	switch ce {
	case BlackAndWhite1:
		return "BlackAndWhite1"
	case Grayscale4:
		return "Grayscale4"
	case Grayscale8:
		return "Grayscale8"
	case Grayscale16:
		return "Grayscale16"
	case RGB24:
		return "RGB24"
	case RGB48:
		return "RGB48"
	case RGBa32:
		return "RGBa32"
	case RGBa64:
		return "RGBa64"
	}

	return "Unknown"
}

// DecodeColorEntry decodes [ColorEntry] out of its XML string representation.
func DecodeColorEntry(s string) ColorEntry {
	switch s {
	case "BlackAndWhite1":
		return BlackAndWhite1
	case "Grayscale4":
		return Grayscale4
	case "Grayscale8":
		return Grayscale8
	case "Grayscale16":
		return Grayscale16
	case "RGB24":
		return RGB24
	case "RGB48":
		return RGB48
	case "RGBa32":
		return RGBa32
	case "RGBa64":
		return RGBa64
	}

	return UnknownColorEntry
}
