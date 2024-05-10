// MFP  - Miulti-Function Printers and scanners toolkit
// argv - Argv parsing mini-library
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Command -- a command definition.

package argv

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Command defines a command.
//
// Every command MUST have a name and MAY have Options,
// positional Parameters and SubCommands
//
// It corresponds to the following usage syntax:
//
//   command [options] [params]
//   command [options] sub-command ...
//
// Parameters and SubCommands are mutually exclusive.
type Command struct {
	// Command name.
	Name string

	// Help string, a single-line description.
	Help string

	// Description contains a long command explanation.
	Description string

	// Options, if any.
	Options []Option

	// Positional parameters, if any.
	Parameters []Parameter

	// Sub-commands, if any.
	SubCommands []Command

	// Handler is called when Command is being invoked.
	// If Handler is nil, DefaultHandler will be used instead.
	Handler func(*Invocation) error
}

// Verify checks correctness of Command definition. It fails if any
// error is found and returns description of the first caught error
func (cmd *Command) Verify() error {
	// Command must have a name
	if cmd.Name == "" {
		return errors.New("missed command name")
	}

	// Parameters and SubCommands are mutually exclusive
	if cmd.hasParameters() && cmd.hasSubCommands() {
		return fmt.Errorf(
			"%s: Parameters and SubCommands are mutually exclusive",
			cmd.Name)
	}

	// Verify Options and Parameters
	err := cmd.verifyOptions()
	if err == nil {
		err = cmd.verifyParameters()
	}

	if err != nil {
		return fmt.Errorf("%s: %s", cmd.Name, err)
	}

	// Verify SubCommands
	err = cmd.verifySubCommands()
	if err != nil {
		return fmt.Errorf("%s.%s", cmd.Name, err)
	}

	return err
}

// verifyOptions verifies command options
func (cmd *Command) verifyOptions() error {
	optnames := make(map[string]struct{})
	for _, opt := range cmd.Options {
		err := opt.verify()
		if err != nil {
			return err
		}

		names := append([]string{opt.Name}, opt.Aliases...)
		for _, name := range names {
			if _, found := optnames[name]; found {
				return fmt.Errorf(
					"duplicated option %q", name)
			}

			optnames[name] = struct{}{}
		}
	}

	return nil
}

// verifyParameters verifies command parameters
func (cmd *Command) verifyParameters() error {
	// Verify each parameter individually
	paramnames := make(map[string]struct{})
	for _, param := range cmd.Parameters {
		err := param.verify()
		if err != nil {
			return err
		}

		if _, found := paramnames[param.Name]; found {
			return fmt.Errorf(
				"duplicated parameter %q", param.Name)
		}

		paramnames[param.Name] = struct{}{}
	}

	// Verify parameters disposition
	var repeated, optional *Parameter

	for i := range cmd.Parameters {
		param := &cmd.Parameters[i]

		if param.optional() {
			if repeated != nil {
				return fmt.Errorf(
					"optional parameter %q used after repeated %q",
					param.Name, repeated.Name)
			}

			optional = param
		} else {
			if optional != nil {
				return fmt.Errorf(
					"required parameter %q used after optional %q",
					param.Name, optional.Name)
			}
		}

		if param.repeated() {
			if repeated != nil {
				return fmt.Errorf(
					"repeated parameter used twice (%q and %q)",
					repeated.Name, param.Name)
			}

			repeated = param
		}
	}

	return nil
}

// verifySubCommands verifies command SubCommands
func (cmd *Command) verifySubCommands() error {
	subcmdnames := make(map[string]struct{})
	for _, subcmd := range cmd.SubCommands {
		if _, found := subcmdnames[subcmd.Name]; found {
			return fmt.Errorf(
				"duplicated subcommand %q", subcmd.Name)
		}

		subcmdnames[subcmd.Name] = struct{}{}

		err := subcmd.Verify()
		if err != nil {
			return err
		}
	}

	return nil
}

// Parse parses Command's arguments and returns either
// Invocation or error.
func (cmd *Command) Parse(argv []string) (*Invocation, error) {
	return cmd.ParseWithParent(nil, argv)
}

// ParseWithParent is like (*Command) Parse(), but allows to specify
// the parent Invocation. It is intended for implementing sub-commands.
func (cmd *Command) ParseWithParent(parent *Invocation,
	argv []string) (*Invocation, error) {
	prs := newParser(cmd, argv)

	return prs.parse(parent)
}

// Run parses the command, then calls its handler
func (cmd *Command) Run(argv []string) error {
	return cmd.RunWithParent(nil, argv)
}

// RunWithParent is like (*Command) Run(), but allows to specify
// the parent Invocation. It is intended for implementing sub-commands.
func (cmd *Command) RunWithParent(parent *Invocation, argv []string) error {
	inv, err := cmd.ParseWithParent(parent, argv)
	if err == nil {
		err = cmd.handler(inv)
	}

	return err
}

// Main emulates main function for the command.
//
// It is intended to implement the body of the
// standard main function:
//
//   // main function for the MyCommand
//   func main() {
//           MyCommand.Main()
//   }
//
// It calls (*Command) Run() passing os.Args as input,
// prints error message, if any, and returns appropriate
// status code to the system.
func (cmd *Command) Main() {
	err := cmd.Run(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

// handler calls cmd.Handler, or DefaultHandler, if
// cmd.Handler is not set.
func (cmd *Command) handler(inv *Invocation) error {
	hnd := DefaultHandler

	switch {
	case inv.immediate != nil:
		hnd = inv.immediate
	case cmd.Handler != nil:
		hnd = cmd.Handler
	}

	return hnd(inv)
}

// Complete returns array of completion suggestions for
// the Command when used with specified (probably incomplete)
// command line.
func (cmd *Command) Complete(cmdline string) []string {
	return nil
}

// hasOptions tells if Command has Options
func (cmd *Command) hasOptions() bool {
	return len(cmd.Options) != 0
}

// hasParameters tells if Command has Parameters
func (cmd *Command) hasParameters() bool {
	return len(cmd.Parameters) != 0
}

// hasSubCommands tells if Command has SubCommands
func (cmd *Command) hasSubCommands() bool {
	return len(cmd.SubCommands) != 0
}

// FindSubCommand finds a Command's SubCommands by name
//
// The name may be abbreviated, and it handles unambiguous
// abbreviation automatically.
//
// If sub-command is not found or ambiguity cannot be resolved,
// it returns nil and appropriate error.
//
// If you want more control, you may want to look to
// the (*Command) FindSubCommandCandidates() function as well.
func (cmd *Command) FindSubCommand(name string) (*Command, error) {
	subcommands := cmd.FindSubCommandCandidates(name)

	switch {
	case len(subcommands) == 0:
		return nil, fmt.Errorf("unknown sub-command: %q", name)
	case len(subcommands) > 1:
		return nil, fmt.Errorf("ambiguous sub-command: %q", name)
	}

	return subcommands[0], nil
}

// FindSubCommandCandidates finds Command's SubCommands by name.
//
// The name may be abbreviated, so in a case of inexact
// match it may return more that one possible candidates.
//
// If no matches found it will return nil and in a case
// of exact match it will return just a single command,
// even if more inexact matches exist
//
// This is up to the caller how to handle this ambiguity.
func (cmd *Command) FindSubCommandCandidates(name string) []*Command {
	var inexact []*Command
	for i := range cmd.SubCommands {
		subcmd := &cmd.SubCommands[i]

		if name == subcmd.Name {
			return []*Command{subcmd}
		}

		if strings.HasPrefix(subcmd.Name, name) {
			inexact = append(inexact, subcmd)
		}
	}

	return inexact
}
