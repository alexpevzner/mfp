// MFP - Miulti-Function Printers and scanners toolkit
// WSD core protocol
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// EndpointReference

package wsd

import (
	"github.com/alexpevzner/mfp/xmldoc"
)

// EndpointReference represents a WSA endpoint address.
type EndpointReference struct {
	Address AnyURI // Endpoint address
}

// DecodeEndpointReference decodes EndpointReference from the XML tree
func DecodeEndpointReference(root xmldoc.Element) (
	ref EndpointReference, err error) {

	defer func() { err = xmlErrWrap(root, err) }()

	Address := xmldoc.Lookup{Name: NsAddressing + ":Address", Required: true}
	missed := root.Lookup(&Address)
	if missed != nil {
		err = xmlErrMissed(missed.Name)
		return
	}

	ref.Address, err = DecodeAnyURI(Address.Elem)

	return
}

// ToXML generates XML tree for the EndpointReference
func (ref EndpointReference) ToXML() xmldoc.Element {
	elm := xmldoc.Element{
		Name: NsAddressing + ":EndpointReference",
		Children: []xmldoc.Element{
			{
				Name: NsAddressing + ":Address",
				Text: string(ref.Address),
			},
		},
	}

	return elm
}
