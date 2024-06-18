// MFP  - Miulti-Function Printers and scanners toolkit
// argv - Argv parsing mini-library
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// (*Command) Main() test

package argv

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// TestMain is a test for (*Command) Main()
func TestMain(t *testing.T) {
	buf := &bytes.Buffer{}
	cmd := Command{
		Name: "test",
		Parameters: []Parameter{
			{Name: "greeting..."},
		},
		Handler: func(inv *Invocation) error {
			buf.WriteString(
				strings.Join(inv.Values("greeting"), ", "))
			return nil
		},
	}

	saveArgs := os.Args
	saveDieOutput := dieOutput
	saveDieExit := dieExit
	defer func() {
		os.Args = saveArgs
		dieOutput = saveDieOutput
		dieExit = saveDieExit
	}()

	os.Args = []string{"test", "hello", "world"}
	cmd.Main()

	expected := "hello, world"
	received := buf.String()

	if expected != received {
		t.Errorf("test 1: expected: `%s`, received: `%s`",
			expected, received)
	}

	os.Args = []string{"test"}
	dieOutput = buf
	dieExit = func(int) {}
	buf.Reset()

	cmd.Main()

	expected = "missed parameter: \"greeting...\"\n"
	received = buf.String()

	if expected != received {
		t.Errorf("test 2: expected: `%s`, received: `%s`",
			expected, received)
	}
}
