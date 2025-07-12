// MFP - Miulti-Function Printers and scanners toolkit
// WS-Scan core protocol
//
// Copyright (C) 2024 and up by Yogesh Singla (yogeshsingla481@gmail.com)
// See LICENSE for license terms and conditions
//
// scanner location

package wsscan

import (
	"github.com/OpenPrinting/go-mfp/util/optional"
	"github.com/OpenPrinting/go-mfp/util/xmldoc"
)

// ScannerLocation holds administratively assigned location information about the scanner.
// The optional ScannerLocation element specifies the administratively assigned location
// of the scanner. The configuration of the ScannerLocation element's value is
// implementation-specific; for example, you can configure this value through the
// scanner's local console or the device's web server. A scan device can return
// multiple versions of this element to enable support for multiple localized languages
// by using the xml:lang attribute.
//
// XML Usage:
//
//	<wscn:ScannerLocation xml:lang="..." lang="xs:string">
//	  text
//	</wscn:ScannerLocation>
//
// Attributes:
//   - lang (xs:string, optional): A character string that identifies the languages
//     of the string that string specifies.
//
// Text value: A character string that specifies the scanner's location.
//
// Parent elements: ScannerDescription
type ScannerLocation struct {
	Location string
	Lang     optional.Val[string]
}

// decodeScannerLocation decodes a [ScannerLocation] from an XML element.
// It extracts the text content and optional xml:lang attribute from the XML element.
// The xml:lang attribute is treated as a single string value.
func decodeScannerLocation(root xmldoc.Element) (sl ScannerLocation, err error) {
	sl.Location = root.Text
	if attr, found := root.AttrByName("xml:lang"); found {
		sl.Lang = optional.New(attr.Value)
	}
	return
}

// toXML converts a [ScannerLocation] to an XML element.
// It creates an XML element with the given name, sets the text content,
// and adds an xml:lang attribute if language information is available.
// The xml:lang attribute is set as a single string value.
func (sl ScannerLocation) toXML(name string) xmldoc.Element {
	elm := xmldoc.Element{Name: name, Text: sl.Location}
	lang := optional.Get(sl.Lang)
	if lang != "" {
		elm.Attrs = []xmldoc.Attr{{Name: "xml:lang", Value: lang}}
	}
	return elm
}
