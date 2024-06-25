// MFP - Miulti-Function Printers and scanners toolkit
// The "cups" command
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Options handling

package cmdcups

import (
	"strings"

	"github.com/alexpevzner/mfp/argv"
	"github.com/alexpevzner/mfp/ipp"
)

// optAttrsGet returns --attrs option (list of requested attributes).
func optAttrsGet(inv *argv.Invocation) (attrs []string) {
	params := inv.Values("--attrs")
	for _, param := range params {
		for _, name := range strings.Split(param, ",") {
			if name != "" {
				attrs = append(attrs, name)
			}
		}
	}

	return
}

// optAttrsGet is the completion callback for the --attrs option.
func optAttrsComplete(arg string) (compl []string, flags argv.CompleterFlags) {
	infos := ((*ipp.PrinterAttributes)(nil)).KnownAttrs()

	attrName := arg
	prefix := ""

	if i := strings.LastIndex(attrName, ","); i >= 0 {
		attrName = arg[i+1:]
		prefix = arg[:i]
	}

	for _, info := range infos {
		if strings.HasPrefix(info.Name, attrName) {
			compl = append(compl, prefix+info.Name)
		}
	}

	return
}
