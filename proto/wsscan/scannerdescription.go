// MFP - Miulti-Function Printers and scanners toolkit
// WS-Scan core protocol
//
// Copyright (C) 2024 and up by Yogesh Singla (yogeshsingla481@gmail.com)
// See LICENSE for license terms and conditions
//
// scanner description

package wsscan

import (
	"github.com/OpenPrinting/go-mfp/util/optional"
	"github.com/OpenPrinting/go-mfp/util/xmldoc"
)

// ScannerDescription holds descriptive information about the scanner.
// The ScannerDescription element contains child elements that provide descriptive
// information about the scanner, including its name, location, and other details.
//
// XML Usage:
//
//	<wscn:ScannerDescription>
//	  child elements
//	</wscn:ScannerDescription>
//
// Attributes: None
//
// Text value: None
//
// Child elements:
//   - ScannerInfo (optional): Administratively assigned descriptive information
//   - ScannerLocation (optional): Administratively assigned location information
//   - ScannerName (required): Administratively assigned user-friendly name
type ScannerDescription struct {
	ScannerInfo     optional.Val[ScannerInfo]
	ScannerLocation optional.Val[ScannerLocation]
	ScannerName     ScannerName
}

// decodeScannerDescription decodes a [ScannerDescription] from an XML element.
// It extracts child elements ScannerName, ScannerInfo, and ScannerLocation from the XML element.
// ScannerName is required, while ScannerInfo and ScannerLocation are optional.
func decodeScannerDescription(root xmldoc.Element) (sd ScannerDescription, err error) {
	defer func() { err = xmldoc.XMLErrWrap(root, err) }()

	// Lookup relevant XML elements
	scannerName := xmldoc.Lookup{Name: NsWSCN + ":ScannerName", Required: true}
	scannerInfo := xmldoc.Lookup{Name: NsWSCN + ":ScannerInfo"}
	scannerLocation := xmldoc.Lookup{Name: NsWSCN + ":ScannerLocation"}

	missed := root.Lookup(&scannerName, &scannerInfo, &scannerLocation)
	if missed != nil {
		err = xmldoc.XMLErrMissed(missed.Name)
		return
	}

	// Decode ScannerName (required)
	sn, err := decodeScannerName(scannerName.Elem)
	if err != nil {
		return sd, err
	}
	sd.ScannerName = sn

	// Decode ScannerInfo (optional)
	if scannerInfo.Found {
		si, err := decodeScannerInfo(scannerInfo.Elem)
		if err != nil {
			return sd, err
		}
		sd.ScannerInfo = optional.New(si)
	}

	// Decode ScannerLocation (optional)
	if scannerLocation.Found {
		sl, err := decodeScannerLocation(scannerLocation.Elem)
		if err != nil {
			return sd, err
		}
		sd.ScannerLocation = optional.New(sl)
	}

	return sd, nil
}

// toXML converts a [ScannerDescription] to an XML element.
// It creates an XML element with the given name and adds child elements
// for ScannerName (required), ScannerInfo (optional), and ScannerLocation (optional).
func (sd ScannerDescription) toXML(name string) xmldoc.Element {
	elm := xmldoc.Element{Name: name}

	// Add ScannerName (required)
	elm.Children = append(elm.Children, sd.ScannerName.toXML(NsWSCN+":ScannerName"))

	// Add ScannerInfo (optional)
	if sd.ScannerInfo != nil {
		info := optional.Get(sd.ScannerInfo)
		elm.Children = append(elm.Children, info.toXML(NsWSCN+":ScannerInfo"))
	}

	// Add ScannerLocation (optional)
	if sd.ScannerLocation != nil {
		location := optional.Get(sd.ScannerLocation)
		elm.Children = append(elm.Children, location.toXML(NsWSCN+":ScannerLocation"))
	}

	return elm
}
