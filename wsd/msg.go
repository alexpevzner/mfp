// MFP - Miulti-Function Printers and scanners toolkit
// WSD core protocol
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Package documentation

package wsd

import (
	"fmt"

	"github.com/alexpevzner/mfp/xmldoc"
)

// Msg represents a WSD protocol message.
type Msg struct {
	Hdr  Hdr  // Message header
	Body Body // Message body
}

// DecodeMsg decodes [msg] from the XML tree
func DecodeMsg(root xmldoc.Element) (m Msg, err error) {
	const (
		rootName = NsSOAP + ":" + "Envelope"
		hdrName  = NsSOAP + ":" + "Header"
		bodyName = NsSOAP + ":" + "Body"
	)

	defer func() { err = xmlErrWrap(root, err) }()

	// Check root element
	if root.Name != rootName {
		err = fmt.Errorf("%s: missed", rootName)
		return
	}

	// Look for Header and Body elements
	hdr := xmldoc.Lookup{Name: hdrName, Required: true}
	body := xmldoc.Lookup{Name: bodyName, Required: true}

	missed := root.Lookup(&hdr, &body)
	if missed != nil {
		err = fmt.Errorf("%s: missed", missed.Name)
		return
	}

	// Decode message header
	m.Hdr, err = DecodeHdr(hdr.Elem)
	if err != nil {
		return
	}

	// Decode message body
	switch m.Hdr.Action {
	case ActHello:
		m.Body, err = DecodeHello(body.Elem)
	case ActBye:
		m.Body, err = DecodeBye(body.Elem)
	default:
		err = fmt.Errorf("%s: unhanded action ", m.Hdr.Action)
	}

	return
}

// ToXML generates XML tree for the message
func (m Msg) ToXML() xmldoc.Element {
	elm := xmldoc.Element{
		Name: NsSOAP + ":" + "Envelope",
		Children: []xmldoc.Element{
			m.Hdr.ToXML(),
			xmldoc.Element{
				Name:     NsSOAP + ":" + "Body",
				Children: []xmldoc.Element{m.Body.ToXML()},
			},
		},
	}

	return elm
}

// Body represents a message body.
type Body interface {
	ToXML() xmldoc.Element
}
