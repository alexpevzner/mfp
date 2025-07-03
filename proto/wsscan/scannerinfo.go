// MFP - Miulti-Function Printers and scanners toolkit
// wsscan core protocol
//
// Copyright (C) 2024 and up by Yogesh Singla (yogeshsingla481@gmail.com)
// See LICENSE for license terms and conditions
//
// Scanner info

package wsscan

import "github.com/OpenPrinting/go-mfp/util/xmldoc"

type ScannerInfo struct {
	Info string
	Lang string
}

func decodeScannerInfo(root xmldoc.Element) (si ScannerInfo, err error) {
	si.Info = root.Text
	if attr, found := root.AttrByName("xml:lang"); found {
		si.Lang = attr.Value
	}
	return
}

func (si ScannerInfo) toXML(name string) xmldoc.Element {
	elm := xmldoc.Element{Name: name, Text: si.Info}
	if si.Lang != "" {
		elm.Attrs = []xmldoc.Attr{{Name: "xml:lang", Value: si.Lang}}
	}
	return elm
}
