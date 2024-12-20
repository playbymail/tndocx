// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package tndocx

import (
	"bytes"
	"regexp"
	"unicode/utf8"
)

const (
	// CR and LF are the ASCII codes for carriage return and line feed, respectively.
	// They are used to represent the end of a line in text files and are needed for
	// cleaning up the text from Windows and MacOS line endings.
	CR = '\r'
	LF = '\n'
)

var (
	rxUnitCourier  = regexp.MustCompile(`^courier \d{4}c\d$`)
	rxUnitElement  = regexp.MustCompile(`^element \d{4}e\d$`)
	rxUnitFleet    = regexp.MustCompile(`^fleet \d{4}f\d$`)
	rxUnitGarrison = regexp.MustCompile(`^garrison \d{4}g\d$`)
	rxUnitTribe    = regexp.MustCompile(`^tribe \d{4}$`)

	rxCourierHeader  = regexp.MustCompile(`^courier \d{4}c\d,`)
	rxElementHeader  = regexp.MustCompile(`^element \d{4}e\d,`)
	rxFleetHeader    = regexp.MustCompile(`^fleet \d{4}f\d,`)
	rxGarrisonHeader = regexp.MustCompile(`^garrison \d{4}g\d,`)
	rxTribeHeader    = regexp.MustCompile(`^tribe \d{4},`)

	rxTurnHeader = regexp.MustCompile(`^current turn \d{3,4}-\d{1,2}\(#\d+\),`)

	rxFleetMovement = regexp.MustCompile(`^(calm|mild|strong|gale) (ne|se|sw|nw|n|s) fleet movement:`)
	rxScoutLine     = regexp.MustCompile(`^scout [1-8]:`)

	rxCourierStatus  = regexp.MustCompile(`^\d{4}c\d status:`)
	rxElementStatus  = regexp.MustCompile(`^\d{4}e\d status:`)
	rxFleetStatus    = regexp.MustCompile(`^\d{4}f\d status:`)
	rxGarrisonStatus = regexp.MustCompile(`^\d{4}g\d status:`)
	rxTribeStatus    = regexp.MustCompile(`^\d{4} status:`)
)

func IsFleetMovement(line []byte) bool {
	return rxFleetMovement.Match(line)
}

func IsMovementLine(line []byte) bool {
	return IsTribeMovement(line) || IsTribeFollows(line) || IsTribeGoesTo(line) || IsScoutLine(line) || IsFleetMovement(line)
}

// IsScoutLine determines if a line represents a TribeNet scout command.
// Example: "scout 1: scout s-pr"
func IsScoutLine(line []byte) bool {
	return rxScoutLine.Match(line)
}

func IsTribeFollows(line []byte) bool {
	return bytes.HasPrefix(line, []byte("tribe follows "))
}

func IsTribeGoesTo(line []byte) bool {
	return bytes.HasPrefix(line, []byte("tribe goes to "))
}

func IsTribeMovement(line []byte) bool {
	return bytes.HasPrefix(line, []byte("tribe movement:"))
}

// IsTurnHeader determines if a line represents a TribeNet turn header.
func IsTurnHeader(line []byte) bool {
	return rxTurnHeader.Match(line)
}

// IsUnitHeader determines if a line represents a TribeNet unit header.
// It checks for five different types of unit headers:
//   - Tribe headers
//   - Courier headers
//   - Element headers
//   - Fleet headers
//   - Garrison headers
//
// Returns true if the line matches any of these header patterns.
func IsUnitHeader(line []byte) bool {
	return rxTribeHeader.Match(line) || rxCourierHeader.Match(line) || rxElementHeader.Match(line) || rxFleetHeader.Match(line) || rxGarrisonHeader.Match(line)
}

// IsUnitStatus determines if a line represents a TribeNet unit status line.
// It checks for five different types of unit status lines:
//   - Tribe status
//   - Courier status
//   - Element status
//   - Fleet status
//   - Garrison status
//
// Returns true if the line matches any of these status line patterns.
func IsUnitStatus(line []byte) bool {
	return rxTribeStatus.Match(line) || rxCourierStatus.Match(line) || rxElementStatus.Match(line) || rxFleetStatus.Match(line) || rxGarrisonStatus.Match(line)
}

// RemoveNonMappingLines filters an input slice of lines, keeping only:
// - Unit headers
// - Turn headers
// - Movement lines
// - Unit status lines
// Returns a new slice containing only the matching lines
func RemoveNonMappingLines(input [][]byte) [][]byte {
	output := make([][]byte, 0, len(input))
	for _, line := range input {
		if IsUnitHeader(line) {
			output = append(output, line)
		} else if IsTurnHeader(line) {
			output = append(output, line)
		} else if IsMovementLine(line) {
			output = append(output, line)
		} else if IsUnitStatus(line) {
			output = append(output, line)
		}
	}
	return output
}

// RemoveLeadingBlankLines trims the leading blank lines from the slice of byte slices.
// Returns the trimmed slice. If the input slice is empty or contains only blank lines,
// returns an empty slice.
func RemoveLeadingBlankLines(lines [][]byte) [][]byte {
	if lines == nil {
		return nil
	}
	for len(lines) != 0 && len(lines[0]) == 0 {
		lines = lines[1:]
	}
	return lines
}

// RemoveTrailingBlankLines trims the trailing blank lines from the slice of byte slices.
// Returns the trimmed slice. If the input contains only blank lines, returns an empty slice.
func RemoveTrailingBlankLines(lines [][]byte) [][]byte {
	if lines == nil {
		return nil
	}
	end := len(lines)
	for end > 0 && len(lines[end-1]) == 0 {
		end--
	}
	return lines[:end]
}

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

type Section struct {
	Id     int
	Header []byte
	Turn   []byte
	Moves  struct {
		Movement []byte
		Follows  []byte
		GoesTo   []byte
		Fleet    []byte
		Scouts   [][]byte
	}
	Status []byte
}

// SectionInput splits the input into lines and assigns lines to their own sections.
// Each element in the input should get a single section
// Each section should contain only movement lines, turn header, and unit header.
func SectionInput(input []byte) (sections []*Section) {
	var section *Section
	for _, line := range bytes.Split(input, []byte{'\n'}) {
		if len(line) == 0 {
			continue
		} else if IsUnitHeader(line) {
			section = &Section{Id: len(sections) + 1, Header: line}
			sections = append(sections, section)
		} else if section == nil {
			continue
		} else if IsFleetMovement(line) {
			section.Moves.Fleet = line
		} else if IsTribeFollows(line) {
			section.Moves.Follows = line
		} else if IsTribeGoesTo(line) {
			section.Moves.GoesTo = line
		} else if IsTribeMovement(line) {
			section.Moves.Movement = line
		} else if IsScoutLine(line) {
			section.Moves.Scouts = append(section.Moves.Scouts, line)
		} else if IsTurnHeader(line) {
			section.Turn = line
		} else if IsUnitStatus(line) {
			section.Status = line
		}
	}
	return sections
}

var (
	// pre-computed lookup table for delimiters
	isSpaceDelimiter [256]bool
)

func init() {
	// initialize the lookup table for delimiters
	for _, ch := range []byte{'\n', ',', '(', ')', '\\', ':'} {
		isSpaceDelimiter[ch] = true
	}
}

// CompressSpaces reduces runs of spaces and tabs to a single space.
// Discards insignificant spaces (for example, before and after delimiters).
// Example: "tribe   0123 ,  ( status ). " -> "tribe 0123,(status)"
func CompressSpaces(input []byte) []byte {
	if len(input) == 0 {
		return input
	}
	output := bytes.NewBuffer(make([]byte, 0, len(input)))
	prevCharWasDelimiter := false
	for len(input) != 0 {
		// if we find a space, advance the input to the end of the run of spaces
		// and decide whether to keep the space or not. if it's insignificant,
		// (meaning it's preceded or followed by a delimiter), discard it.
		if input[0] == ' ' || input[0] == '\t' { // found a space
			// advance input to the end of the run of spaces
			for len(input) != 0 && (input[0] == ' ' || input[0] == '\t') {
				input = input[1:]
			}
			// next character is a delimiter if is end-of-input or a delimiter.
			nextCharIsDelimiter := len(input) == 0 || isSpaceDelimiter[input[0]]
			// discard this run of spaces if they are preceded or followed by a delimiter
			if prevCharWasDelimiter || nextCharIsDelimiter {
				continue
			}
			output.WriteByte(' ')
			continue
		}
		output.WriteByte(input[0])
		prevCharWasDelimiter = isSpaceDelimiter[input[0]]
		input = input[1:]
	}
	return output.Bytes()
}

var (
	reBackslashDash = regexp.MustCompile(`\\+-+ *`)

	reBackslashComma = regexp.MustCompile(`\\+,+`)
	reBackslashUnit  = regexp.MustCompile(`\\+(\d{4}(?:[cefg]\d)?)`)
	reCommaBackslash = regexp.MustCompile(`,+\\`)
	reDirectionUnit  = regexp.MustCompile(`\b(ne|se|sw|nw|n|s) (\d{4}(?:[cefg]\d)?)`)

	reRunOfBackslashes = regexp.MustCompile(`\\\\+`)
	reRunOfComma       = regexp.MustCompile(`,,+`)
)

// PreProcessMovementLine processes a movement line to fix issues with backslash or direction followed by a unit ID.
// Caller must have already compressed spaces on input line.
func PreProcessMovementLine(line []byte) []byte {
	// replace backslash+dash with backslash
	line = reBackslashDash.ReplaceAll(line, []byte{'\\'})

	// replace backslash+comma and comma+backslash with backslash
	line = reBackslashComma.ReplaceAll(line, []byte{'\\'})
	line = reCommaBackslash.ReplaceAll(line, []byte{'\\'})

	// fix issues with backslash or direction followed by a unit ID
	line = reBackslashUnit.ReplaceAll(line, []byte{',', '$', '1'})
	line = reDirectionUnit.ReplaceAll(line, []byte{'$', '1', ',', '$', '2'})

	// reduce runs of certain punctuation to a single punctuation character
	line = reRunOfBackslashes.ReplaceAll(line, []byte{'\\'})
	line = reRunOfComma.ReplaceAll(line, []byte{','})

	// tweak the fleet movement to remove the trailing comma from the observations
	line = bytes.ReplaceAll(line, []byte{',', ')'}, []byte{')'})

	// remove all trailing backslashes from the line
	line = bytes.TrimRight(line, "\\")

	return line
}
