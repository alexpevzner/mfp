// MFP - Miulti-Function Printers and scanners toolkit
// XML mini library
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// XML decoder

package xml

import (
	"bytes"
	"encoding/xml"
	"io"
)

// Decode parses XML document, and represents it as a linear
// sequence of XML elements
//
// Each element has a Name, Path, which is a full path to the element,
// starting from root and Text, which is XML element body, stripped
// from leading and trailing space, and Children, which includes
// its direct children, children of children and so on.
//
// Path uses '/' character as a path separator and starts with '/'
//
// Decoded elements are ordered according to their layout in the
// input document (i.e., root, then first child of the root, then
// its first child and so on), and [Element.Children] are ordered
// the same way.
//
// Namespace prefixes are rewritten according to the 'ns' map.
// Full namespace URL used as map index, and value that corresponds
// to the index replaced with map value. If URL is not found in the
// map, prefix replaced with "-" string
func Decode(ns Namespace, in io.Reader) ([]*Element, error) {
	var elements []*Element
	var elem *Element
	var path bytes.Buffer

	decoder := xml.NewDecoder(in)
	for {
		token, err := decoder.Token()
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			break
		}

		switch t := token.(type) {
		case xml.StartElement:
			// Decode name and path.
			// Namespace translation is handled here.
			var name string
			if t.Name.Space != "" {
				var ok bool
				name, ok = ns.ByURL(t.Name.Space)
				if !ok {
					name = "-"
				}
			}

			if name != "" {
				name += ":"
			}
			name += t.Name.Local

			path.WriteByte('/')
			path.WriteString(name)

			// Create an element
			elem = &Element{
				Name:   name,
				Path:   path.String(),
				Parent: elem,
			}
			elements = append(elements, elem)

			for p := elem.Parent; p != nil; p = p.Parent {
				p.Children = append(p.Children, elem)
			}

		case xml.EndElement:
			elem = elem.Parent
			if elem != nil {
				path.Truncate(len(elem.Path))
			} else {
				path.Truncate(0)
			}

		case xml.CharData:
			if elem != nil {
				elem.Text = string(bytes.TrimSpace(t))
			}
		}
	}

	return elements, nil
}
