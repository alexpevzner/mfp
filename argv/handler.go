// MFP  - Miulti-Function Printers and scanners toolkit
// argv - Argv parsing mini-library
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// DefaultHandler for Command

package argv

import (
	"fmt"
	"strings"
)

// DefaultHandler is the default Command's Handler
func DefaultHandler(inv *Invocation) error {
	subcmd, subargv := inv.SubCommand()
	if subcmd != nil {
		return subcmd.RunWithParent(inv, subargv)
	}

	argv := append([]string{inv.Cmd().Name}, inv.Argv()...)
	return fmt.Errorf("unhandled command: %s", strings.Join(argv, " "))
}
