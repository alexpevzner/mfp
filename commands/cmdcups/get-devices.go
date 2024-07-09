// MFP - Miulti-Function Printers and scanners toolkit
// The "cups" command
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// The "get-printers" command.

package cmdcups

import (
	"context"

	"github.com/alexpevzner/mfp/argv"
	"github.com/alexpevzner/mfp/cups"
	"github.com/alexpevzner/mfp/env"
)

// cmdGetPrinters defines the "get-printers" sub-command.
var cmdGetDevices = argv.Command{
	Name:    "get-devices",
	Help:    "Search for available devices",
	Handler: cmdGetDevicesHandler,
	Options: []argv.Option{
		optLimit,
		argv.HelpOption,
	},
}

// cmdGetPrintersHandler is the "get-printers" command handler
func cmdGetDevicesHandler(ctx context.Context, inv *argv.Invocation) error {
	// Prepare arguments
	dest := optCUPSURL(inv)

	sel := &cups.GetDevicesSelection{
		Limit: optLimitGet(inv),
	}

	// Perform the query
	pager := env.NewPager()

	pager.Printf("CUPS: %s", dest)

	clnt := cups.NewClient(dest, nil)
	devices, err := clnt.CUPSGetDevices(ctx, sel, []string{"all"})
	if err != nil {
		return err
	}

	for _, dev := range devices {
		pager.Printf("")
		devAttrsFormat(pager, dev)
	}

	return pager.Display()
}
