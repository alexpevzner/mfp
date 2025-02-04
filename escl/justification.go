// MFP - Miulti-Function Printers and scanners toolkit
// eSCL core protocol
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// ADF image justification.

package escl

import "github.com/alexpevzner/mfp/xmldoc"

// Justification specifies how the ADF justify the document.
type Justification int

// Known CCD Channels.
const (
	UnknownJustification Justification = iota // Unknown CCD
	Left                                      // Left (horizontal)
	Right                                     // Right (horizontal)
	Top                                       // Top (vertical)
	Bottom                                    // Bottom (vertical)
	Center                                    // Center (both)
)

// decodeJustification decodes [Justification] from the XML tree.
func decodeJustification(root xmldoc.Element) (jst Justification, err error) {
	return decodeEnum(root, DecodeJustification, NsScan)
}

// toXML generates XML tree for the [Justification].
func (jst Justification) toXML(name string) xmldoc.Element {
	return xmldoc.Element{
		Name: name,
		Text: NsScan + ":" + jst.String(),
	}
}

// String returns a string representation of the [Justification]
func (jst Justification) String() string {
	switch jst {
	case Left:
		return "Left"
	case Right:
		return "Right"
	case Top:
		return "Top"
	case Bottom:
		return "Bottom"
	case Center:
		return "Center"
	}

	return "Unknown"
}

// DecodeJustification decodes [Justification] out of its XML
// string representation.
func DecodeJustification(s string) Justification {
	switch s {
	case "Left":
		return Left
	case "Right":
		return Right
	case "Top":
		return Top
	case "Bottom":
		return Bottom
	case "Center":
		return Center
	}

	return UnknownJustification
}
