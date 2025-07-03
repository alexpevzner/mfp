// MFP - Miulti-Function Printers and scanners toolkit
// wsscan core protocol
//
// Copyright (C) 2024 and up by Yogesh Singla (yogeshsingla481@gmail.com)
// See LICENSE for license terms and conditions
//
// Common types and functions for tests

package wsscan

import (
	"reflect"
	"strings"
	"testing"

	"github.com/OpenPrinting/go-mfp/util/generic"
	"github.com/OpenPrinting/go-mfp/util/xmldoc"
)

// testEnumType is the common interface of all enum-alike types
type testEnumType interface {
	~int
	String() string
	toXML(name string) xmldoc.Element
}

// testEnum defines a test vector for enum-alike type
type testEnum[T testEnumType] struct {
	decodeStr func(string) T                  // Decode value from string
	decodeXML func(xmldoc.Element) (T, error) // Decode from XML element
	dataset   []testEnumData[T]               // Test data cases
}

// testEnumData represents a test data entry for enum-like types,
// like ColorMode etc
type testEnumData[T testEnumType] struct {
	v T      // enum value
	s string // string representation
}

// run performs tests
func (test testEnum[T]) run(t *testing.T) {
	const xmlName = "test:elem"

	typeName := reflect.TypeOf(T(0)).String()
	if i := strings.LastIndexByte(typeName, '.'); i >= 0 {
		typeName = typeName[i+1:]
	}

	withUnknown := generic.CopySlice(test.dataset)
	withUnknown = append(withUnknown, testEnumData[T]{0, "Unknown"})

	// Test T.String()
	for _, data := range withUnknown {
		s := data.v.String()
		if s != data.s {
			t.Errorf("%s(%d).String():\n"+
				"expected: %q\n"+
				"present:  %q\n",
				typeName, data.v,
				data.s, s)
		}

	}

	// Test T.toXML()
	for _, data := range test.dataset {
		xml := data.v.toXML(xmlName)
		exp := xmldoc.Element{
			Name: xmlName,
			Text: data.v.String(),
		}

		if !reflect.DeepEqual(xml, exp) {
			t.Errorf("%s.toXML():\n"+
				"expected: %s\n"+
				"present:  %s\n",
				data.v,
				exp.EncodeString(nil),
				xml.EncodeString(nil))
		}
	}

	// test decodeStr
	for _, data := range withUnknown {
		v := test.decodeStr(data.s)
		if v != data.v {
			t.Errorf("Decode%s(%q):\n"+
				"expected: %s\n"+
				"present:  %s\n",
				typeName, data.s, data.v, v)
		}
	}

	// test decodeXML
	for _, data := range test.dataset {
		xml := xmldoc.Element{
			Name: xmlName,
			Text: data.s,
		}

		// normal decode
		v, err := test.decodeXML(xml)
		if err != nil {
			t.Errorf("decode%s():\n"+
				"input: %s\n"+
				"error: %q\n",
				typeName, xml.EncodeString(nil), err)
			continue
		}

		if v != data.v {
			t.Errorf("decode%s():\n"+
				"input:    %s\n"+
				"expected: %s\n"+
				"present:  %s\n",
				typeName, xml.EncodeString(nil), data.v, v)
		}

		// invalid value
		xml.Text = data.s + "-invalid"

		_, err = test.decodeXML(xml)
		if err == nil {
			t.Errorf("decode%s():\n"+
				"input: %s\n"+
				"error: expected but did'n occur",
				typeName, xml.EncodeString(nil))
		}
	}
}
