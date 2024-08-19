// MFP - Miulti-Function Printers and scanners toolkit
// Device discovery
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Scanner input source

package discovery

// InputSource defines input sources, supported by scanner
type InputSource int

// InputSource bits:
const (
	InputPlaten InputSource = 1 << iota // Platen source
	InputADF                            // Automatic Document Feeder
)
