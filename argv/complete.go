// MFP  - Miulti-Function Printers and scanners toolkit
// argv - Argv parsing mini-library
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Value completers.

package argv

import (
	"io/fs"
	"strings"
)

// Completer is a callback called for auto-completion
//
// Any Option or Parameter may have its own Completer.
//
// It receives the Option's value prefix, already typed
// by user, and must return a slice of completion candidates
// that match the prefix.
//
// For example, if possible Option or Parameter values are "Richard",
// "Roger" and  "Robert", then, depending of supplied prefix, the following
// output is expected:
//
//   "R"   -> ["Richard", "Roger", "Robert"]
//   "Ro"  -> ["Roger", "Robert"]
//   "Rog" -> ["Roger"]
//   "Rol" -> []
type Completer func(string) ([]string, CompleterFlags)

// CompleterFlags returned as a second return value from Completer
// and provides some hints how caller should interpret returned
// completion candidates.
//
// See each flag's documentation for more details.
type CompleterFlags int

const (
	// CompleterNoSpace indicates that caller should not
	// append a space after completion.
	//
	// This is useful, for example, for file path completion
	// that ends with path separator character (i.e., '/')
	// and user is prompted to continue entering a full
	// file name.
	CompleterNoSpace CompleterFlags = 1 << iota
)

// CompleteStrings returns a completer, that performs auto-completion,
// choosing from a set of supplied strings.
func CompleteStrings(s []string) Completer {
	// Create a copy of input, to protect from callers
	// that may change the slice after the call.
	set := make([]string, len(s))
	copy(set, s)

	// Create completer
	return func(in string) ([]string, CompleterFlags) {
		out := []string{}
		for _, member := range set {
			if len(in) < len(member) &&
				strings.HasPrefix(member, in) {
				out = append(out, member)
			}
		}
		return out, 0
	}
}

// CompleteFs returns a completer, that performs file name auto-completion
// on a top of a virtual (or real) filesystem, represented as fs.FS,
//
// getwd callback returns a current directory within that file system.
// It's signature is compatible with os.Getwd(), so this function can
// be used directly.
//
// If getwd is nil, current directory assumed to be "/"
func CompleteFs(fsys fs.FS, getwd func() (string, error)) Completer {
	fscompl := newFscompleter(fsys, getwd)
	return func(arg string) ([]string, CompleterFlags) {
		return fscompl.complete(arg)
	}
}
