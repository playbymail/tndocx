// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package docx provides a reader for Microsoft Word documents.
package docx

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"os"
	"regexp"
)

// ReadFile loads a Word document from a file, converts it to plain text, and returns the text as a byte slice.
func ReadFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ReadBuffer(data)
}

// ReadBuffer loads a Word document from a byte slice, converts it to plain text, and returns the text as a byte slice.
func ReadBuffer(data []byte) ([]byte, error) {
	return Read(bytes.NewReader(data))
}

// Read reads a Word document, converts it to plain text, and returns the text as a byte slice.
func Read(r *bytes.Reader) ([]byte, error) {
	zr, err := zip.NewReader(r, r.Size())
	if err != nil {
		return nil, err
	}

	doc := docx{
		zipFileReader: nil,
		Files:         zr.File,
		FilesContent:  map[string][]byte{},
	}

	for _, f := range doc.Files {
		contents, _ := doc.retrieveFileContents(f.Name)
		doc.FilesContent[f.Name] = contents
	}

	// convert the xml data to a slice of word tokens
	doc.listP(string(doc.FilesContent["word/document.xml"]))

	// convert the word tokens to a slice containing all the words.
	// we collapse spaces into a single space and can't tell the difference between a space and a tab.
	// we also destroy all the original Word tables.
	result := &bytes.Buffer{}
	for _, word := range doc.WordsList {
		for column, content := range word.Content {
			if column != 0 {
				result.WriteString(" ")
			}
			result.WriteString(content)
		}
		result.WriteByte('\n')
	}

	return result.Bytes(), nil
}

// http://officeopenxml.com/anatomyofOOXML.php

// copied from https://github.com/lu4p/cat/blob/master/docxtxt/docxreader.go
// and licensed as provided in the COPYING file in this folder.

// docx zip struct
type docx struct {
	zipFileReader *zip.ReadCloser
	Files         []*zip.File
	FilesContent  map[string][]byte
	WordsList     []*words
}

type words struct {
	Content []string
}

// read all files contents
func (d *docx) retrieveFileContents(filename string) ([]byte, error) {
	var file *zip.File
	for _, f := range d.Files {
		if f.Name == filename {
			file = f
		}
	}

	if file == nil {
		return nil, errors.New(filename + " file not found")
	}

	reader, err := file.Open()
	if err != nil {
		return nil, err
	}

	return io.ReadAll(reader)
}

var (
	rxRunT = regexp.MustCompile(`(?U)(<w:r>|<w:r .*>)(.*)(</w:r>)`)
	rxT    = regexp.MustCompile(`(?U)(<w:t>|<w:t .*>)(.*)(</w:t>)`)
)

// get w:t value
func (d *docx) getT(item string) {
	var subStr string
	data := item
	w := new(words)
	content := []string{}

	wrMatch := rxRunT.FindAllStringSubmatchIndex(data, -1)
	// loop r
	for _, rMatch := range wrMatch {
		rData := data[rMatch[4]:rMatch[5]]
		wtMatch := rxT.FindAllStringSubmatchIndex(rData, -1)
		for _, match := range wtMatch {
			subStr = rData[match[4]:match[5]]
			content = append(content, subStr)
		}
	}
	w.Content = content
	d.WordsList = append(d.WordsList, w)
}

var (
	rxP = regexp.MustCompile(`(?U)<w:p[^>]*>(.*)</w:p>`)
)

// hasP identify the paragraph
func hasP(data string) bool {
	result := rxP.MatchString(data)
	return result
}

var (
	rxListP = regexp.MustCompile(`(?U)<w:p[^>]*(.*)</w:p>`)
)

// listP for w:p tag value
func (d *docx) listP(data string) {
	var result []string
	for _, match := range rxListP.FindAllStringSubmatch(data, -1) {
		result = append(result, match[1])
	}
	for _, item := range result {
		if hasP(item) {
			d.listP(item)
			continue
		}
		d.getT(item)
	}
}
