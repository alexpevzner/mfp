// MFP - Miulti-Function Printers and scanners toolkit
// eSCL core protocol
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Common type for Range of some value.

package escl

import (
	"strconv"

	"github.com/alexpevzner/mfp/optional"
	"github.com/alexpevzner/mfp/xmldoc"
)

// Range commonly used to specify the range of some parameter, like
// brightness, contrast etc.
type Range struct {
	Min    int               // Minimal supported value
	Max    int               // Maximal supported value
	Normal int               // Normal value
	Step   optional.Val[int] // Step between the subsequent values
}

// decodeRange decodes [Range] from the XML tree
func decodeRange(root xmldoc.Element) (r Range, err error) {
	defer func() { err = xmldoc.XMLErrWrap(root, err) }()

	// Lookup message elements
	min := xmldoc.Lookup{Name: NsScan + ":Min", Required: true}
	max := xmldoc.Lookup{Name: NsScan + ":Max", Required: true}
	normal := xmldoc.Lookup{Name: NsScan + ":Normal", Required: true}
	step := xmldoc.Lookup{Name: NsScan + ":Step"}

	missed := root.Lookup(&min, &max, &normal, &step)
	if missed != nil {
		err = xmldoc.XMLErrMissed(missed.Name)
		return
	}

	// Decode elements
	r.Min, err = decodeNonNegativeInt(min.Elem)
	if err == nil {
		r.Min, err = decodeNonNegativeInt(min.Elem)
	}
	if err == nil {
		r.Max, err = decodeNonNegativeInt(max.Elem)
	}
	if err == nil {
		r.Normal, err = decodeNonNegativeInt(normal.Elem)
	}
	if err == nil && step.Found {
		var tmp int
		tmp, err = decodeNonNegativeInt(step.Elem)
		r.Step = optional.New(tmp)
	}

	return
}

// ToXML generates XML tree for the [Range].
func (r Range) ToXML(name string) xmldoc.Element {
	elm := xmldoc.Element{
		Name: name,
		Children: []xmldoc.Element{
			{
				Name: NsScan + ":" + "Min",
				Text: strconv.FormatUint(uint64(r.Min), 10),
			},
			{
				Name: NsScan + ":" + "Max",
				Text: strconv.FormatUint(uint64(r.Max), 10),
			},
			{
				Name: NsScan + ":" + "Normal",
				Text: strconv.FormatUint(uint64(r.Normal), 10),
			},
		},
	}

	if r.Step != nil {
		step := xmldoc.Element{
			Name: NsScan + ":" + "Step",
			Text: strconv.FormatUint(uint64(*r.Step), 10),
		}
		elm.Children = append(elm.Children, step)
	}

	return elm
}
