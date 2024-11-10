// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package docx

import "bytes"

// WordDocType represents the type of Word document.
type WordDocType int

const (
	Unknown WordDocType = iota
	Doc                 // Word 97–2003 Documents
	Docx                // Word 2007 and Later Documents
)

var (
	docMagicNumber  = []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}
	docxMagicNumber = []byte{0x50, 0x4B, 0x03, 0x04}
)

// DetectWordDocType checks the initial bytes of a file to determine if it's a Word document.
// Note: this is not a 100% accurate method since DOCX files share the same magic number as ZIP files.
func DetectWordDocType(data []byte) WordDocType {
	if bytes.HasPrefix(data, docMagicNumber) {
		return Doc
	} else if bytes.HasPrefix(data, docxMagicNumber) {
		return Docx
	}
	return Unknown
}
