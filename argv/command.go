// MFP  - Miulti-Function Printers and scanners toolkit
// argv - Argv parsing mini-library
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Command definition structures.

package argv

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
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
}

// Option defines an option.
//
// Option MUST have a name and MAY have one or more aliases.
//
// Name and Aliases of all Options MUST be unique within a scope
// of Command that defines them (sub-commands have their own scopes).
//
// Option MAY have a value. Presence of name is indicated by the
// non-nil Validate field.
//
// Option may have either short or long syntax:
//
//   -c                                 - short option without value
//   --long-name                        - long option without value
//   -c XXX or -cXXX                    - short name with value
//   --long-name XXX or --long-name=XXX - long name with value
//
// Short options without value can be combined:
//
//   -cru equals to -c -r -u
//
// Short name starts with a single dash (-) character followed
// by a single alphanumeric character.
//
// Long name starts with a double dash (--) characters followed
// by an alphanumeric character, optionally followed by a sequence
// of characters, that include only alphanumeric characters and dashes.
//
//   -x            - valid
//   -abc          - invalid; short option with a long name
//   --x           - valid (the long option, though name is 1-character)
//   --long        - valid
//   --long-option - valid
//   ---long       - invalid; character after -- must be alphanumerical
//
// This naming convention is consistent with GNU extensions to the POSIX
// recommendations for command-line options:
//
//   https://www.gnu.org/software/libc/manual/html_node/Argument-Syntax.html
type Option struct {
	// Name is the option name.
	Name string

	// Aliases are the option aliases, if any.
	Aliases []string

	// Help string, a single-line description.
	Help string

	// Conflicts, if not nit, contains names of other Options
	// that MUST NOT be used together with this option.
	Conflicts []string

	// Requires, if not nil, contains names of other Options,
	// that MUST be used together with this option.
	Requires []string

	// Validate callback called to validate parameter.
	//
	// Use nil to indicate that this option has no value.
	Validate func(string) error

	// Complete callback called for auto-completion.
	// It receives the prefix, already typed by user
	// (which may be empty) and must return completion
	// suggestions without that prefix.
	Complete func(string) []string
}

// Parameter defines a positional parameter.
//
// Parameter MUST have a name, and names of all Parameters
// MUST be unique within a scope  of Command that defines them
// (sub-commands have their own scopes).
//
// Parameter names used to generate help messages and to
// access parameters by name, hence requirement of uniqueness.
//
// If name of the Parameter ends with ellipsis (...), this is
// repeated parameter:
//
//   copy source... destination
//
// If name of the Parameter is taken into square braces ([name]),
// this is optional parameter:
//
//   print document [format]
//
// Optional parameter may be omitted.
//
// Ellipses a square braces syntax may be combined:
//
//   list [file...]
//
// Non-optional repeated parameter will consume 1 or more
// parameter values. Optional repeated parameter will consume
// 0 or more parameter values.
//
// Not all combinations of required, optional and repeated
// parameters are valid.
//
// Valid combinations:
//
//   cmd param1 param2 [param3] [param4]      - OK
//   cmd param1 param2 [param3] [param4...]   - OK
//   cmd param1 param2 param3... param4       - OK
//   cmd param1 param2... param3 param4       - OK
//
// Inlaid combinations:
//
//   Required parameter       cmd param1 [param2] param3
//   can't follow optional
//   parameter
//
//   Optional parameter       cmd param1 param2... [param3]
//   can't follow repeated    cmd param1 [param2...] [param3]
//   parameter
//
//   Only one repeated        cmd param1 param2... param3...
//   parameter is allowed
//
// These rules exist so simplify unambiguous matching of actual
// parameters against formal (declared) ones.
type Parameter struct {
	// Name is the parameter name.
	Name string

	// Help string, a single-line description.
	Help string

	// Validate callback called to validate parameter
	Validate func(string) error

	// Complete callback called for auto-completion.
	// It receives the prefix, already typed by user
	// (which may be empty) and must return completion
	// suggestions without that prefix.
	Complete func(string) []string
}

// Action defines action to be taken when Command is
// applied to the command line.
type Action struct {
	options map[string][]string
}

// ----- Command methods -----

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

// Apply applies Command to argument. On success
// it returns Action which defines further processing.
func (cmd *Command) Apply(argv []string) (*Action, error) {
	prs := newParser(cmd, argv)
	err := prs.parse()

	return nil, err
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

// ----- Option methods -----

// verify checks correctness of Option definition. It fails if any
// error is found and returns description of the first caught error
func (opt *Option) verify() error {
	// Option must have a name
	if opt.Name == "" {
		return errors.New("option must have a name")
	}

	// Verify syntax of option Name and Aliases
	names := append([]string{opt.Name}, opt.Aliases...)
	for _, name := range names {
		err := opt.verifyNameSyntax(name)
		if err != nil {
			return err
		}
	}

	// Verify name syntax of Conflicts and Requires
	for _, name := range opt.Conflicts {
		err := opt.verifyNameSyntax(name)
		if err != nil {
			return fmt.Errorf("Conflicts: %w", err)
		}
	}

	for _, name := range opt.Requires {
		err := opt.verifyNameSyntax(name)
		if err != nil {
			return fmt.Errorf("Requires: %w", err)
		}
	}

	return nil
}

// verifyNameSyntax verifies option name syntax
func (opt *Option) verifyNameSyntax(name string) error {
	var check string
	var short bool

	switch {
	case strings.HasPrefix(name, "--"):
		check = name[2:]
	case strings.HasPrefix(name, "-"):
		short = true
		check = name[1:]

	default:
		return fmt.Errorf(
			"option must start with dash (-): %q", name)
	}

	if check == "" {
		return fmt.Errorf("empty option name: %q", name)
	}

	if c := nameCheck(check); c >= 0 {
		return fmt.Errorf(
			"invalid char '%c' in option: %q", c, name)
	}

	if short && len(check) > 1 {
		return fmt.Errorf(
			"short option with long name: %q", name)
	}

	return nil
}

// withValue tells if Option has a value
func (opt *Option) withValue() bool {
	return opt.Validate != nil
}

// complete is the convenience wrapper around Option.Complete
// callback. It call callback only if one is not nil.
func (opt *Option) complete(prefix string) (compl []string) {
	if opt.Complete != nil {
		compl = opt.Complete(prefix)
	}
	return
}

// ----- Parameter methods -----

// verify checks correctness of Parameter definition. It fails if any
// error is found and returns description of the first caught error
func (param *Parameter) verify() error {
	// Parameter must have a name
	if param.Name == "" {
		return errors.New("parameter must have a name")
	}

	// Verify name syntax
	check := param.Name
	if strings.HasPrefix(check, "[") {
		// If name starts with "[", this is optional parameter,
		// and it must end with "]"
		if strings.HasSuffix(check, "]") {
			check = check[1 : len(check)-1]
		} else {
			err := fmt.Errorf("missed closing ']' character in parameter %q",
				param.Name)
			return err
		}
	}

	if strings.HasSuffix(check, "...") {
		// Strip trailing "...", if any
		check = check[0 : len(check)-3]
	}

	// Check remaining name
	if check == "" {
		return fmt.Errorf("parameter name is empty: %q", param.Name)
	}

	if c := nameCheck(check); c >= 0 {
		return fmt.Errorf("invalid char '%c' in parameter: %q",
			c, param.Name)
	}

	return nil
}

// optional returns true if parameter is required
func (param *Parameter) required() bool {
	return !param.optional()
}

// optional returns true if parameter is optional
func (param *Parameter) optional() bool {
	return strings.HasPrefix(param.Name, "[")
}

// repeated returns true if parameter is repeated
func (param *Parameter) repeated() bool {
	return strings.HasSuffix(param.Name, "...") ||
		strings.HasSuffix(param.Name, "...]")
}

// complete is the convenience wrapper around Parameter.Complete
// callback. It call callback only if one is not nil.
func (param *Parameter) complete(prefix string) (compl []string) {
	if param.Complete != nil {
		compl = param.Complete(prefix)
	}
	return
}

// ----- Action methods -----

// Getopt returns value of option or parameter as a single string.
//
// For multi-value options and repeated parameters values
// are concatenated into the single string using CSV encoding.
func (act *Action) Getopt(name string) (val string, found bool) {
	return "", false
}

// GetoptSlice returns value of option or parameter as a slice of string.
// If option is not found, it returns nil
func (act *Action) GetoptSlice(name string) (val []string) {
	return nil
}

// ----- Miscellaneous functions -----

// nameCheck function verifies syntax of Options and
// Parameters names. Valid name starts with letter or
// digit and then consist of letters, digits and dash
// characters.
//
// It returns the first invalid character, if one is
// encountered, or -1 otherwise.
func nameCheck(name string) rune {
	for i, c := range name {
		switch {
		// Letters and digits always allowed
		case unicode.IsLetter(c) || unicode.IsDigit(c):

		// Dash allowed expect the very first character
		case i > 0 && c == '-':

		// Other characters not allowed
		default:
			return c
		}
	}

	return -1
}
