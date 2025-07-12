// MFP - Miulti-Function Printers and scanners toolkit
// WS-Scan core protocol
//
// Copyright (C) 2024 and up by Yogesh Singla (yogeshsingla481@gmail.com)
// See LICENSE for license terms and conditions
//
// Basic XML decoding functions

package wsscan

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/OpenPrinting/go-mfp/util/optional"
	"github.com/OpenPrinting/go-mfp/util/xmldoc"
)

// decodeInt decodes integer from the XML tree.
func decodeInt(root xmldoc.Element) (v int, err error) {
	var v64 int64
	v64, err = strconv.ParseInt(root.Text, 10, 64)

	switch {
	case err != nil:
		err = fmt.Errorf("invalid int: %q", root.Text)
	case v64 < math.MinInt32 || v64 > math.MaxInt32:
		err = fmt.Errorf("int out of range: %d", v64)
	}

	if err != nil {
		err = xmldoc.XMLErrWrap(root, err)
		return 0, err
	}

	return int(v64), nil
}

// decodeNonNegativeInt decodes non-negative integer from the XML tree.
func decodeNonNegativeInt(root xmldoc.Element) (v int, err error) {
	var v64 int64
	v64, err = strconv.ParseInt(root.Text, 10, 64)

	switch {
	case err != nil:
		err = fmt.Errorf("invalid int: %q", root.Text)
	case v64 < 0 || v64 > math.MaxInt32:
		err = fmt.Errorf("int out of range: %d", v64)
	}

	if err != nil {
		err = xmldoc.XMLErrWrap(root, err)
		return 0, err
	}

	return int(v64), nil
}

// decodeBool decodes boolean from the XML tree.
func decodeBool(root xmldoc.Element) (v bool, err error) {
	switch root.Text {
	case "true":
		return true, nil
	case "false":
		return false, nil
	}

	err = fmt.Errorf("invalid bool: %q", root.Text)
	err = xmldoc.XMLErrWrap(root, err)

	return
}

// decodeNMTOKEN decodes xs:NMTOKEN from the XML tree.
//
// XML 1.0 defines xs:NMTOKEN as a token, composed of characters,
// digits, “.”, “:”, “-”, and the characters defined by Unicode,
// such as “combining” or “extender”.
//
// Here we implement the simplified version that only allows
// Latin characters and punctuation signs mentioned above.
// This simplification looks reasonable, as we are implementing
// eSCL parser, not the universal XML toolkit.
func decodeNMTOKEN(root xmldoc.Element) (v string, err error) {
	if root.Text != "" {
		for _, c := range root.Text {
			switch {
			case '0' <= c && c <= '9':
			case 'a' <= c && c <= 'z':
			case 'A' <= c && c <= 'Z':
			case c == '.' || c == ':' || c == '-':
			default:
				goto ERROR
			}
		}

		return root.Text, nil
	}

ERROR:
	err = fmt.Errorf("invalid xs:NMTOKEN: %q", root.Text)
	err = xmldoc.XMLErrWrap(root, err)

	return
}

// decodeEnum decodes value of enum-alike type T from the XML tree
//
// decode is the type-specific function that decodes T from string
// (i.e., DecodeColorEntry for ColorEntry).
func decodeEnum[T ~int](root xmldoc.Element,
	decode func(string) T) (val T, err error) {

	val = decode(root.Text)
	if val != 0 {
		return
	}

	typeName := reflect.TypeOf(T(0)).String()
	if i := strings.LastIndexByte(typeName, '.'); i >= 0 {
		typeName = typeName[i+1:]
	}

	err = fmt.Errorf("invalid %s: %q", typeName, root.Text)
	err = xmldoc.XMLErrWrap(root, err)

	return
}

// decodeOptional wraps decodeXXX function that decodes xmldoc.Element
// into the value of type T, to decode xmldoc.Element into the optional
// value of type optional.Val[T]
func decodeOptional[T any](root xmldoc.Element,
	decodeXXX func(xmldoc.Element) (T, error)) (optional.Val[T], error) {

	v, err := decodeXXX(root)
	if err != nil {
		return nil, err
	}

	return optional.New(v), nil
}
