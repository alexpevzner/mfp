// MFP - Miulti-Function Printers and scanners toolkit
// eSCL core protocol
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Job state

package escl

import "github.com/alexpevzner/mfp/xmldoc"

// JobState represents the Job current state
type JobState int

// Known Job states
const (
	UnknownJobState JobState = iota // Unknown Job state
	Canceled                        // Job was canceled by user
	Aborted                         // Job was aborted due to fatal error
	Completed                       // Job is finished successfully
	Pending                         // Job was initiated
	Processing                      // Job is in progress
)

// decodeJobState decodes [JobState] from the XML tree.
func decodeJobState(root xmldoc.Element) (state JobState, err error) {
	return decodeEnum(root, DecodeJobState)
}

// toXML generates XML tree for the [JobState].
func (state JobState) toXML(name string) xmldoc.Element {
	return xmldoc.Element{
		Name: name,
		Text: state.String(),
	}
}

// String returns a string representation of the [JobState]
func (state JobState) String() string {
	switch state {
	case Canceled:
		return "Canceled"
	case Aborted:
		return "Aborted"
	case Completed:
		return "Completed"
	case Pending:
		return "Pending"
	case Processing:
		return "Processing"
	}

	return "Unknown"
}

// DecodeJobState decodes [JobState] out of its XML string representation.
func DecodeJobState(s string) JobState {
	switch s {
	case "Canceled":
		return Canceled
	case "Aborted":
		return Aborted
	case "Completed":
		return Completed
	case "Pending":
		return Pending
	case "Processing":
		return Processing
	}

	return UnknownJobState
}
