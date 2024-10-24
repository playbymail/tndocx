// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package tndocx

import (
	"bytes"
	"unicode/utf8"
)

const (
	// CR and LF are the ASCII codes for carriage return and line feed, respectively.
	// They are used to represent the end of a line in text files and are needed for
	// cleaning up the text from Windows and Mac OS line endings.
	CR = '\r'
	LF = '\n'
)

var (
	// convert RuneError to a string and then to a []byte
	runeErrorByte = []byte(string(utf8.RuneError))
)

// ScrubBadUTF8 processes a byte slice and replaces any invalid UTF-8 sequences
// with the UTF-8 replacement character. Returns a new byte slice containing
// valid UTF-8 sequences.
func ScrubBadUTF8(input []byte) []byte {
	if input == nil {
		return nil
	}

	output := bytes.NewBuffer(make([]byte, 0, len(input)))
	for len(input) != 0 {
		r, w := utf8.DecodeRune(input)
		if r == utf8.RuneError {
			output.Write(runeErrorByte)
			continue
		}
		output.Write(input[:w])
		input = input[w:]
	}
	return output.Bytes()
}

// ScrubEOL converts different types of EOL to Unix EOL.
// Converts Windows EOL (CR+LF) to Unix EOL (LF).
// Converts Classic Mac EOL (CR) to Unix EOL (LF).
// Unix EOL (LF) passes through unchanged.
func ScrubEOL(input []byte) []byte {
	if len(input) == 0 {
		return input
	}
	output := bytes.NewBuffer(make([]byte, 0, len(input)))
	for len(input) != 0 {
		if input[0] == CR { // window or maybe classic mac
			input = input[1:]
			// found CR, check for CR LF
			if len(input) != 0 && input[0] == LF {
				input = input[1:]
			}
			output.WriteByte(LF)
			continue
		}
		output.WriteByte(input[0])
		input = input[1:]
	}
	return output.Bytes()
}

// TrimLeadingBlankLines trims the leading blank lines from the slice of byte slices.
// Returns the trimmed slice. If the input slice is empty or contains only blank lines,
// returns an empty slice.
func TrimLeadingBlankLines(lines [][]byte) [][]byte {
	if lines == nil {
		return nil
	}
	for len(lines) != 0 && len(lines[0]) == 0 {
		lines = lines[1:]
	}
	return lines
}

// TrimTrailingBlankLines trims the trailing blank lines from the slice of byte slices.
// Returns the trimmed slice. If the input contains only blank lines, returns an empty slice.
func TrimTrailingBlankLines(lines [][]byte) [][]byte {
	if lines == nil {
		return nil
	}
	end := len(lines)
	for end > 0 && len(lines[end-1]) == 0 {
		end--
	}
	return lines[:end]
}
