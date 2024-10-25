// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package tndocx_test

import (
	"github.com/playbymail/tndocx"
	"testing"
)

func TestCompressSpaces(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "no spaces",
			input:    "tribe0123",
			expected: "tribe0123",
		},
		{
			name:     "single spaces",
			input:    "tribe 0123",
			expected: "tribe 0123",
		},
		{
			name:     "multiple spaces",
			input:    "tribe   0123",
			expected: "tribe 0123",
		},
		{
			name:     "spaces after delimiter",
			input:    "tribe,   0123",
			expected: "tribe,0123",
		},
		{
			name:     "spaces at end of input",
			input:    "tribe,   0123  ",
			expected: "tribe,0123",
		},
		{
			name:     "tabs and spaces",
			input:    "tribe\t \t0123",
			expected: "tribe 0123",
		},
		{
			name:     "complex example",
			input:    "tribe   0123,   status:  active  ( good )",
			expected: "tribe 0123,status:active(good)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(tndocx.CompressSpaces([]byte(tt.input)))
			if got != tt.expected {
				t.Errorf("CompressSpaces() = %q, want %q", got, tt.expected)
			}
		})
	}
}
