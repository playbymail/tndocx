// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package tndocx

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// types in an unparsed report file

type Report struct {
	FileName string           `json:"file-name"`
	TurnId   string           `json:"turn-id"`
	Units    map[string]*Unit `json:"units,omitempty"`
	Meta     struct {
		GeneratedBy string `json:"generated-by"`
		Version     string `json:"version,omitempty"`
		Timestamp   int64  `json:"timestamp,omitempty"`
	} `json:"metadata"`
}

type Units []*Unit

type Unit struct {
	Id     string   `json:"id"`
	Input  string   `json:"text,omitempty"`
	Name   string   `json:"name,omitempty"` // in Access, this is the TribeName field
	From   string   `json:"from,omitempty"`
	To     string   `json:"to,omitempty"`
	Winds  *Winds   `json:"winds,omitempty"`
	Moves  []*Step  `json:"moves,omitempty"`
	Scouts []*Scout `json:"scouts,omitempty"`
	Status string   `json:"status,omitempty"`
}

type Winds struct {
	Strength  string `json:"strength,omitempty"`
	Direction string `json:"direction,omitempty"`
}

type Step struct {
	Follows      string `json:"follows,omitempty"`
	GoesTo       string `json:"goes-to,omitempty"`
	Step         string `json:"step,omitempty"`
	Still        bool   `json:"still,omitempty"`
	Observations string `json:"observations,omitempty"`
}

type Scout struct {
	Id     string   `json:"id"`
	Patrol []string `json:"scout,omitempty"`
	Still  bool     `json:"still,omitempty"`
}

var (
	// rxFleetMovementLine captures fleet movement lines.
	rxFleetMovementLine = regexp.MustCompile(`^(calm|mild|strong|gale) (ne|se|sw|nw|n|s) fleet movement:move(.*)$`)
	rxFleetObservation  = regexp.MustCompile(`^\([^)]*\)`)

	// rxScoutPatrolLine captures scout patrol lines.
	rxScoutPatrolLine = regexp.MustCompile(`^scout ([1-8]):scout(.*)$`)

	// rxTurnHeaderLine is the regular expression that matches the turn header line.
	// that line looks like: "tribe 0138,current hex = ## 0709,(previous hex = ## 0709)"
	rxTribeHeaderLine     = regexp.MustCompile(`^(?:courier|element|garrison|fleet|tribe) (\d{4}(?:[cdefg]\d)?),current hex = (n/a|(?:##|[a-z]{2}) \d{4}),\(previous hex = (n/a|(?:##|[a-z]{2}) \d{4})\)$`)
	rxTribeHeaderMiscLine = regexp.MustCompile(`^(?:courier|element|garrison|fleet|tribe) (\d{4}(?:[cdefg]\d)?),([^,]*),current hex = (n/a|(?:##|[a-z]{2}) \d{4}),\(previous hex = (n/a|(?:##|[a-z]{2}) \d{4})\)$`)

	// rxTribeFollows captures tribe follows lines.
	// these look like:
	// - tribe follows 0987g1
	rxTribeFollowsLine = regexp.MustCompile(`^tribe follows (\d{4}(?:[cdefg]\d)?)$`)

	// rxTribeGoesTo captures tribe goes to lines.
	// these look like:
	// - tribe goes to QQ 0707
	rxTribeGoesToLine = regexp.MustCompile(`^tribe goes to ([a-z][a-z] \d{4})$`)

	// rxTribeMovementLine captures tribe movement lines.
	// these look like:
	// - tribe movement:move
	// - tribe movement:move ne-pr\n-pr,o nw
	// 0987/data/input/0900-09.0987.report.txt:Tribe Movement: Move S-GH,  L NE,  SE,  S\SW-GH,  L SE,  S\SW-GH,  L SE,  S\SW-GH,  L SE,  S\SW-GH,  L SE,  S\SW-PR,  L SE\S-GH,  L NE, River SE S\No Ford on River to SE of HEX
	// - tribe movement:move nw-pr,river sw,ford s,dowdy holler,0987g1\not enough m.p's to move to n into swamp
	rxTribeMovementLine = regexp.MustCompile(`^tribe movement:move(.*)$`)

	// rxTribeStatusLine captures tribe status lines.
	// these look like:
	// - unit status: terrain, settlement, resources, edges, neighboring-terrains, units, maybe-some-other-stuff
	// - 0987 status:grassy hills,dowdy holler,coal,river n ne,ford se s,0987,0987e1
	// - 0987g1 status:conifer hills,west harbor,iron ore,o ne,n,ford se,s,stone road ne n,0987g1
	rxTribeStatusLine = regexp.MustCompile(`\d{4}(?:[cdefg]\d)? status:(.*)$`)

	// - current turn 900-04(#4),summer,fine
	rxTurnHeaderLine = regexp.MustCompile(`^current turn (\d{3,4})-(\d{1,2})`)
)

// ToReport filters an input slice of lines, keeping only:
// - Unit headers
// - Turn headers
// - Movement lines
// - Unit status lines
// Returns a Report containing only the lines needed for mapping.
func ToReport(filename string, input [][]byte) *Report {
	report := &Report{
		FileName: filename,
		Units:    make(map[string]*Unit),
	}
	report.Meta.GeneratedBy = "tn3"
	report.Meta.Version = version.String()
	report.Meta.Timestamp = time.Now().UTC().Unix()
	unit := &Unit{}
	for n, line := range input {
		if match := rxTribeHeaderLine.FindSubmatch(line); match != nil {
			unit = &Unit{
				Id:   string(match[1]),
				From: string(match[3]),
				To:   string(match[2]),
			}
			report.Units[unit.Id] = unit
		} else if match := rxTribeHeaderMiscLine.FindSubmatch(line); match != nil {
			unit = &Unit{
				Id:   string(match[1]),
				Name: string(match[2]),
				From: string(match[4]),
				To:   string(match[3]),
			}
			report.Units[unit.Id] = unit
		} else if IsUnitHeader(line) {
			// this match seems redundant, but it's not.
			// it allows us to capture unit headers that are slightly off.
			// if we didn't, then it would be much harder for the players to debug their reports.
			unit = &Unit{
				Id:    fmt.Sprintf("unit-%03d", n+1),
				Input: string(line),
			}
			report.Units[unit.Id] = unit
		} else if match := rxTurnHeaderLine.FindSubmatch(line); match != nil {
			year, _ := strconv.Atoi(string(match[1]))
			month, _ := strconv.Atoi(string(match[2]))
			report.TurnId = fmt.Sprintf("%04d-%02d", year, month)
		} else if rxTurnHeader.Match(line) {
			// this match seems redundant, but it's not.
			// it allows us to capture turn headers that are slightly off.
			// if we didn't, then it would be much harder for the players to debug their reports.
			report.TurnId = string(line)
		} else if match := rxScoutPatrolLine.FindSubmatch(line); match != nil {
			scout := &Scout{
				Id: string(match[1]),
			}
			for _, step := range strings.Split(string(match[2]), "\\") {
				step = strings.TrimSpace(strings.TrimLeft(strings.TrimRight(step, ", "), ", "))
				if step == "" {
					continue
				}
				scout.Patrol = append(scout.Patrol, step)
			}
			unit.Scouts = append(unit.Scouts, scout)
		} else if match := rxTribeMovementLine.FindSubmatch(line); match != nil {
			for _, step := range strings.Split(string(match[1]), "\\") {
				if step = strings.TrimSpace(step); step == "" {
					continue
				}
				unit.Moves = append(unit.Moves, &Step{
					Step: step,
				})
			}
		} else if match := rxTribeFollowsLine.FindSubmatch(line); match != nil {
			unit.Moves = append(unit.Moves, &Step{Follows: string(match[1])})
		} else if match := rxTribeGoesToLine.FindSubmatch(line); match != nil {
			unit.Moves = append(unit.Moves, &Step{GoesTo: string(match[1])})
		} else if match := rxFleetMovementLine.FindSubmatch(line); match != nil {
			unit.Winds = &Winds{
				Strength:  string(match[1]),
				Direction: string(match[2]),
			}
			for _, step := range strings.Split(string(match[3]), "\\") {
				if step = strings.TrimSpace(step); step == "" {
					continue
				}
				fs := &Step{}
				if shtep, shobvs, ok := strings.Cut(step, "-("); !ok {
					fs.Step = step
				} else {
					fs.Step = strings.TrimSpace(strings.TrimRight(shtep, ","))
					fs.Observations = "(" + strings.TrimSpace(shobvs)
				}
				unit.Moves = append(unit.Moves, fs)
			}
		} else if match := rxTribeStatusLine.FindSubmatch(line); match != nil {
			unit.Status = string(match[1])
		}
	}
	return report
}
