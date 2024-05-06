// MFP  - Miulti-Function Printers and scanners toolkit
// argv - Argv parsing mini-library
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Argv tokenizer

package argv

import (
	"errors"
	"unicode"
)

// Tokenize splits command line string into separate arguments.
//
// It understands the following syntax:
//
//   param1 param2 param3          -> ["param1", "param2", "param3"]
//   param1 "param 2" "param3"     -> ["param1", "param 2", "param3"]
//   param1 hel"lo wo"rld "param3" -> ["param1", "hello world", "param3"]
//
// It recognizes the following C-like escapes within the quoted string:
//
//   "\a" -> 0x07 (audible bell)
//   "\b" -> 0x08 (backspace)
//   "\f" -> 0x0c (form feed - new page)
//   "\n" -> 0x0a (line feed - new line)
//   "\r" -> 0x0d (carriage return)
//   "\t" -> 0x09 (horizontal tab)
//   "\v" -> 0x0b (vertical tab)
//
//   \0   -> 0000
//   \1   -> 0001
//   \12  -> 0012
//   \123 -> 0123
//
//   \x12 -> 0x00
//
// If quoted string syntactically incorrect (for example, quoted string
// is not terminated), the problem will be reported using second (error)
// return value, but it will do it best to pick up tokens correctly; this
// is useful for auto-completion.
func Tokenize(line string) ([]string, error) {
	type tkState int
	const (
		tkSpace   tkState = iota
		tkWord            // Got non-space
		tkQuote           // Got "
		tkQuoteBs         // Got " ... \
		tkHex1            // Got " ... \x
		tkHex2            // Got " ... \xN
		tkOct1            // Got " ... \N
		tkOct2            // Got " ... \NN
	)

	state := tkSpace
	token := ""
	acc := 0
	tokens := []string{}

	for _, c := range line {

		switch state {
		case tkSpace, tkWord:
			if c == '"' {
				state = tkQuote
			} else if unicode.IsSpace(c) {
				if state != tkSpace {
					tokens = append(tokens, token)
					token = ""
					state = tkSpace
				}
			} else {
				state = tkWord
				token += string(c)
			}

		case tkQuote:
			if c == '\\' {
				state = tkQuoteBs
			} else if c == '"' {
				state = tkWord
			} else {
				token += string(c)
			}

		case tkQuoteBs:
			switch c {
			case 'x', 'X':
				acc = 0
				state = tkHex1

			case '0', '1', '2', '3', '4', '5', '6', '7':
				acc = int(c - '0')
				state = tkOct1

			case 'a':
				token += "\a"

			case 'b':
				token += "\b"

			case 'f':
				token += "\f"

			case 'n':
				token += "\n"

			case 'r':
				token += "\r"

			case 't':
				token += "\t"

			case 'v':
				token += "\v"

			default:
				token += string(c)
			}

			if state == tkQuoteBs {
				state = tkQuote
			}

		case tkHex1, tkHex2:
			if n := hexadecimal(c); n >= 0 {
				acc = (acc << 4) | n
				if state == tkHex1 {
					state = tkHex2
				} else {
					token += string([]byte{byte(acc)})
					state = tkQuote
				}
			} else {
				token += string([]byte{byte(acc)})
				if c == '"' {
					state = tkWord
				} else {
					token += string(c)
					state = tkQuote
				}
			}

		case tkOct1, tkOct2:
			if n := octal(c); n >= 0 {
				acc = (acc << 3) | n
				if state == tkOct1 {
					state = tkOct2
				} else {
					token += string([]byte{byte(acc)})
					state = tkQuote
				}
			} else {
				token += string([]byte{byte(acc)})
				if c == '"' {
					state = tkWord
				} else {
					token += string(c)
					state = tkQuote
				}
			}
		}
	}

	switch state {
	case tkWord:
		tokens = append(tokens, token)

	case tkQuote, tkQuoteBs, tkHex1, tkHex2, tkOct1, tkOct2:
		return nil, errors.New("unterminated string")
	}

	return tokens, nil
}

// octal returns numerical value of octal digit c.
// If c is not octal digit, it returns -1.
func octal(c rune) int {
	if '0' <= c && c <= '9' {
		return int(c - '0')
	}
	return -1
}

// hexadecimal returns numerical value of hexadecimal digit c.
// If c is not hexadecimal digit, it returns -1.
func hexadecimal(c rune) int {
	switch {
	case '0' <= c && c <= '9':
		return int(c - '0')
	case 'a' <= c && c <= 'f':
		return int(c-'a') + 10
	case 'A' <= c && c <= 'F':
		return int(c-'A') + 10
	}
	return -1
}
