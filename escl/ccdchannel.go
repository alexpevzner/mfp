// MFP - Miulti-Function Printers and scanners toolkit
// eSCL core protocol
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Scan CCD color channel

package escl

import "github.com/alexpevzner/mfp/xmldoc"

// CcdChannel specifies which CCD color channel to use for grayscale
// and monochrome scannig.
type CcdChannel int

// Known CCD Channels.
const (
	UnknownCcdChannel CcdChannel = iota // Unknown CCD
	Red                                 // Use the RED DDC
	Green                               // Use the Green CCD
	Blue                                // Use the Blue CCD
	NTSC                                // NTSC-standard mix
	GrayCcd                             // Dedicated hardware Gray CCD
	GrayCcdEmulated                     // Emulated Gray CCD (1/3 RGB)
)

// decodeCcdChannel decodes [CcdChannel] from the XML tree.
func decodeCcdChannel(root xmldoc.Element) (ccd CcdChannel, err error) {
	return decodeEnum(root, DecodeCcdChannel)
}

// toXML generates XML tree for the [CcdChannel].
func (ccd CcdChannel) toXML(name string) xmldoc.Element {
	return xmldoc.Element{
		Name: name,
		Text: ccd.String(),
	}
}

// String returns a string representation of the [CcdChannel]
func (ccd CcdChannel) String() string {
	switch ccd {
	case Red:
		return "Red"
	case Green:
		return "Green"
	case Blue:
		return "Blue"
	case NTSC:
		return "NTSC"
	case GrayCcd:
		return "GrayCcd"
	case GrayCcdEmulated:
		return "GrayCcdEmulated"
	}

	return "Unknown"
}

// DecodeCcdChannel decodes [CcdChannel] out of its XML string representation.
func DecodeCcdChannel(s string) CcdChannel {
	switch s {
	case "Red":
		return Red
	case "Green":
		return Green
	case "Blue":
		return Blue
	case "NTSC":
		return NTSC
	case "GrayCcd":
		return GrayCcd
	case "GrayCcdEmulated":
		return GrayCcdEmulated
	}

	return UnknownCcdChannel
}
