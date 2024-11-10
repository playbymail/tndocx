// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package tndocx

import (
	"bytes"
	"errors"
	"github.com/playbymail/tndocx/docx"
	"regexp"
)

func ParseSections(input []byte) ([]*Section, error) {
	if len(input) == 0 {
		return nil, ErrEmptyInput
	}
	sections, err := ParseDocx(input)
	if err != nil && errors.Is(ErrUnknownFormat, err) {
		sections, err = ParseText(input)
	}
	return sections, err
}

func isascii(b byte) bool {
	return 0 < b && b <= 127
}

func ParseDocx(input []byte) ([]*Section, error) {
	if docx.DetectWordDocType(input) != docx.Docx {
		return nil, ErrUnknownFormat
	}

	// extract the text from the Word document
	text, err := docx.ReadBuffer(input)
	if err != nil {
		return nil, err
	}

	return ParseText(text)
}

func ParseText(input []byte) ([]*Section, error) {
	if !(len(input) > 3 && isascii(input[0]) && isascii(input[1]) && isascii(input[2])) {
		return nil, ErrUnknownFormat
	}
	// bug: have to force the entire file to lower case
	input = bytes.ToLower(input)

	// compress spaces within the input
	input = CompressSpaces(input)

	sections := SectionInput(input)
	//log.Printf("sections %8d bytes into %d sections\n", len(input), len(sections))
	for _, section := range sections {
		section.Moves.Movement = scrubMovementLine(section.Moves.Movement)
		//if len(section.Moves.Movement) != 0 {
		//	log.Printf("section %3d: %s\n", section.Id, section.Moves.Movement)
		//}
		section.Moves.Follows = scrubFollowsLine(section.Moves.Follows)
		section.Moves.GoesTo = scrubGoesToLine(section.Moves.GoesTo)
		section.Moves.Fleet = scrubFleetLine(section.Moves.Fleet)
		for n, line := range section.Moves.Scouts {
			section.Moves.Scouts[n] = scrubScoutLine(line)
			//log.Printf("section %3d: %s\n", section.Id, section.Moves.Scouts[n])
		}
		section.Status = scrubStatusLine(section.Status)
		//if len(section.Status) != 0 {
		//	log.Printf("section %3d: %s\n", section.Id, section.Status)
		//}
	}

	return sections, nil
}

// scrubFleetLine does some pre-processing on the fleet line.
func scrubFleetLine(line []byte) []byte {
	if len(line) == 0 {
		return line
	}

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

// scrubFollowsLine does some pre-processing on the follows line.
func scrubFollowsLine(line []byte) []byte {
	if len(line) == 0 {
		return line
	}
	// remove the "tribe follows" prefix
	line = bytes.TrimSpace(bytes.TrimPrefix(line, []byte("tribe follows")))
	// return the pre-processed line
	return line
}

// scrubGoesToLine does some pre-processing on the goes to line.
func scrubGoesToLine(line []byte) []byte {
	if len(line) == 0 {
		return line
	}
	// remove the "tribe goes to" prefix
	line = bytes.TrimSpace(bytes.TrimPrefix(line, []byte("tribe goes to")))
	// return the pre-processed line
	return line
}

// scrubMovementLine does some pre-processing on the movement line.
func scrubMovementLine(line []byte) []byte {
	if len(line) == 0 {
		return line
	}

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

	line = scrubStepResults(line)

	// remove all trailing backslashes from the line
	line = bytes.TrimRight(line, "\\")

	return line
}

// scrubScoutLine does some pre-processing on the scout line.
func scrubScoutLine(line []byte) []byte {
	if len(line) == 0 {
		return line
	}

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

	line = scrubStepResults(line)

	// remove all trailing backslashes from the line
	line = bytes.TrimRight(line, "\\")

	return line
}

// scrubStatusLine does some pre-processing on the status line.
func scrubStatusLine(line []byte) []byte {
	if len(line) == 0 {
		return line
	}

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

	line = scrubStepResults(line)

	// remove all trailing backslashes from the line
	line = bytes.TrimRight(line, "\\")

	return line
}

// scrubStepResults processes a step to normalize commas and spaces around direction codes and unit IDs
//
// while len(step) != 0
//
//	if step does not start with a comma
//	   advance step to the next character
//	else if step starts with an edge code (hsm, lcm, ljm, lsm)
//	     advance step past the edge code
//	     while step starts with a comma or space followed by a direction code (n, s, ne, se, nw, sw)
//	           replace the comma with a space
//	           skip step past the space and the direction code
//	else if step starts with a unit ID
//	     advance past the unit ID
//	     while step starts with a comma or space followed by a unit ID
//	           replace the comma with a space
//	           skip step past the space and the unit ID
//	else
//	   advance skip to the next character
//
// return the line
func scrubStepResults(line []byte) []byte {
	step := line
	// advance to the first comma
	for len(step) > 0 && step[0] != ',' {
		step = step[1:]
	}

	// process all the results
	for len(step) > 0 {
		// does step start with an edge code?
		if match := edgeCodePattern.Find(step); match != nil {
			step = step[len(match):] // advance past the initial edge code

			// process list of direction codes
			for elementMatch := listDirectionPattern.Find(step); elementMatch != nil; elementMatch = listDirectionPattern.Find(step) {
				step[0] = ' '                   // replace comma with space
				step = step[len(elementMatch):] // skip past the matched direction code
			}
			continue
		}

		// does step starts with a unit ID?
		if match := unitIDPattern.Find(step); match != nil {
			step = step[len(match):] // advance past the initial unit ID

			// process list of unit IDs
			for elementMatch := listUnitIDPattern.Find(step); elementMatch != nil; elementMatch = listUnitIDPattern.Find(step) {
				step[0] = ' '                   // replace comma with space
				step = step[len(elementMatch):] // skip past the matched unit ID
			}
			continue
		}

		// no matches, so advance to the next comma
		step = step[1:]
		for len(step) != 0 && step[0] != ',' {
			step = step[1:]
		}
	}

	return line
}

var (
	// Regular expressions for edge codes, unit IDs, and lists of directions and units
	edgeCodePattern      = regexp.MustCompile(`^,(hsm|l|lcm|ljm|lsm|o)\b`)
	unitIDPattern        = regexp.MustCompile(`^,\d{4}([cefg]\d)?\b`)
	listDirectionPattern = regexp.MustCompile(`^[,\s]([ns][ew]?)\b`)
	listUnitIDPattern    = regexp.MustCompile(`^[,\s]\d{4}([cefg]\d)?\b`)
)
