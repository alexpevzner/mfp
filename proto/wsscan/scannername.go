// MFP - Miulti-Function Printers and scanners toolkit
// WS-Scan core protocol
//
// Copyright (C) 2024 and up by Yogesh Singla (yogeshsingla481@gmail.com)
// See LICENSE for license terms and conditions
//
// scanner name

package wsscan

import (
	"github.com/OpenPrinting/go-mfp/util/optional"
	"github.com/OpenPrinting/go-mfp/util/xmldoc"
)

// ScannerName holds the administratively assigned user-friendly name of the scanner.
// The required ScannerName element specifies the administratively assigned user-friendly
// name of the scanner. The configuration of the ScannerName element's value is
// implementation-specific; for example, you can configure this value through the
// scanner's local console or the device's web server. If a device has only one
// hosted service, its friendly name and ScannerName element should have the same value.
// If the device contains several hosted services, ScannerName should identify the scanner.
//
// A scan device can return multiple versions of this element to enable support for
// multiple localized languages by using the xml:lang attribute.
//
// XML Usage:
//
//	<wscn:ScannerName xml:lang="..." lang="xs:string">
//	  text
//	</wscn:ScannerName>
//
// Attributes:
//   - lang (xs:string, optional): A character string that identifies the languages
//     of the string that string specifies.
//
// Text value: A character string that specifies the scanner's user-friendly name.
//
// Parent elements: ScannerDescription
type ScannerName struct {
	Name string
	Lang optional.Val[string]
}

// decodeScannerName decodes a [ScannerName] from an XML element.
// It extracts the text content and optional xml:lang attribute from the XML element.
// The xml:lang attribute is treated as a single string value.
func decodeScannerName(root xmldoc.Element) (sn ScannerName, err error) {
	sn.Name = root.Text
	if attr, found := root.AttrByName("xml:lang"); found {
		sn.Lang = optional.New(attr.Value)
	}
	return
}

// toXML converts a [ScannerName] to an XML element.
// It creates an XML element with the given name, sets the text content,
// and adds an xml:lang attribute if language information is available.
// The xml:lang attribute is set as a single string value.
func (sn ScannerName) toXML(name string) xmldoc.Element {
	elm := xmldoc.Element{Name: name, Text: sn.Name}
	lang := optional.Get(sn.Lang)
	if lang != "" {
		elm.Attrs = []xmldoc.Attr{{Name: "xml:lang", Value: lang}}
	}
	return elm
}
