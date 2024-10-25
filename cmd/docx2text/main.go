// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"bytes"
	"github.com/playbymail/tndocx"
	"github.com/playbymail/tndocx/docx"
	"iter"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func main() {
	log.SetFlags(log.Lshortfile)

	root := "../userdata"
	for clan := range Clans(root) {
		docxPath := filepath.Join(root, clan, "docx")
		for turnId := range TurnIds(docxPath) {
			for reportName := range TurnReports(docxPath, turnId, "docx") {
				started := time.Now()
				// load the Word document
				docxPath := filepath.Join(root, clan, "docx", reportName)
				input, err := os.ReadFile(docxPath)
				if err != nil {
					log.Fatalf("error: %v\n", err)
				}
				log.Printf("%s: %s: loaded  %s\t in %v\n", clan, turnId, docxPath, time.Since(started))

				// extract the text from the Word document
				text, err := docx.ReadBuffer(input)
				if err != nil {
					log.Fatalf("error: %v\n", err)
				}
				log.Printf("%s: %s: read    %s\t in %v\n", clan, turnId, docxPath, time.Since(started))

				// compress spaces within the text
				text = tndocx.CompressSpaces(text)
				log.Printf("%s: %s: despace %s\t in %v\n", clan, turnId, docxPath, time.Since(started))

				// remove unnecessary lines from the text
				lines := bytes.Split(text, []byte{'\n'})
				log.Printf("%s: %s: split   %s\t in %v\n", clan, turnId, docxPath, time.Since(started))
				lines = tndocx.RemoveNonMappingLines(lines)
				log.Printf("%s: %s: trimmed %s\t in %v\n", clan, turnId, docxPath, time.Since(started))

				// write the text to a file
				text = bytes.Join(lines, []byte{'\n'})
				log.Printf("%s: %s: merged  %s\t in %v\n", clan, turnId, docxPath, time.Since(started))

				textPath := strings.TrimSuffix(docxPath, filepath.Ext(docxPath)) + ".txt"
				if err := os.WriteFile(textPath, text, 0644); err != nil {
					log.Fatalf("error: %v\n", err)
				}
				log.Printf("%s: %s: created %s\t in %v\n", clan, turnId, textPath, time.Since(started))
			}
		}
	}
}

var ( // compile the regex patterns
	rxClanId     = regexp.MustCompile(`^0\d\d\d$`)
	rxTurnReport = regexp.MustCompile(`^(\d+-\d+)\.(\d{4}).report\.(docx|txt)$`)
)

// Clans returns an iterator that yields the names of all the clans in the given path.
// The Clan name is the name of the folder and must be a 4-digit number starting with 0.
// Panics if there are any errors reading the path because I don't know how to handle errors in an iterator.
func Clans(path string) iter.Seq[string] {
	// get list of all folders and files in the path
	list, err := os.ReadDir(path)
	if err != nil {
		panic(err)
	}

	// return an iterator that yields the names of all the clans
	return func(yield func(string) bool) {
		for _, item := range list {
			// check if it's a directory and matches pattern
			if !item.IsDir() {
				continue
			} else if match := rxClanId.FindStringSubmatch(item.Name()); match == nil {
				continue
			}
			if !yield(item.Name()) {
				// caller has stopped iteration, so clean up and exit
				return
			}
		}
	}
}

// TurnIds returns an iterator that yields the turn id of all the turn reports in the given path.
// Assumes that we're interested only in the Word documents.
// Panics if there are any errors reading the path because I don't know how to handle errors in an iterator.
func TurnIds(path string) iter.Seq[string] {
	// get list of all folders and files in the path
	list, err := os.ReadDir(path)
	if err != nil {
		panic(err)
	}

	// return an iterator that yields the turn id from all the report documents
	return func(yield func(string) bool) {
		for _, item := range list {
			matches := rxTurnReport.FindStringSubmatch(item.Name())
			if matches == nil {
				continue
			}
			turnId := matches[1]
			if !yield(turnId) {
				// caller has stopped iteration, so clean up and exit
				return
			}
		}
	}
}

// TurnReports returns an iterator that yields the name of all the turn reports in the given path.
// The caller must specify either the Word document or the text document.
// Panics if there are any errors reading the path because I don't know how to handle errors in an iterator.
func TurnReports(path string, turnId, ext string) iter.Seq[string] {
	// get list of all folders and files in the path
	list, err := os.ReadDir(path)
	if err != nil {
		panic(err)
	}

	// return an iterator that yields the turn id from all the report documents
	return func(yield func(string) bool) {
		for _, item := range list {
			matches := rxTurnReport.FindStringSubmatch(item.Name())
			//log.Printf("matches: %d %+v\n", len(matches), matches)
			if matches == nil {
				continue
			} else if !(turnId == matches[1] && ext == matches[3]) {
				//log.Printf("matches: (%q %q) (%q %q)\n", turnId, matches[1], ext, matches[3])
				continue
			}
			reportName := item.Name()
			if !yield(reportName) {
				// caller has stopped iteration, so clean up and exit
				return
			}
		}
	}
}
