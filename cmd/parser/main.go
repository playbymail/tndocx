// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"github.com/playbymail/tndocx"
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

	root, rootStarted := "data/input", time.Now()
	files, err := os.ReadDir(root)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}
	numberOfReportFiles, numberOfTextFiles, numberOfWordFiles := 0, 0, 0
	for _, file := range files {
		started := time.Now()
		if file.IsDir() {
			continue
		}
		if strings.HasSuffix(file.Name(), ".docx") {
			numberOfWordFiles++
		} else if strings.HasSuffix(file.Name(), ".txt") {
			numberOfTextFiles++
		} else {
			continue
		}
		numberOfReportFiles++
		fileName := file.Name()
		filePath := filepath.Join(root, fileName)
		// load the document
		input, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatalf("error: %v\n", err)
		}
		// parse the document
		sections, err := tndocx.ParseSections(input)
		log.Printf("%s: parsed %3d sections in %v\n", fileName, len(sections), time.Since(started))
	}
	log.Printf("parsed text %3d: word %3d: total %3d files in %v\n", numberOfTextFiles, numberOfWordFiles, numberOfReportFiles, time.Since(rootStarted))
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
