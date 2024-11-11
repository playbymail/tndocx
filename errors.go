// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package tndocx

type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrEmptyInput           = Error("empty input")
	ErrInvalidElementId     = Error("invalid element id")
	ErrMissingElementHeader = Error("missing element header")
	ErrMissingField         = Error("missing field")
	ErrNotImplemented       = Error("not implemented")
	ErrUnexpectedInput      = Error("unexpected input")
	ErrUnknownFormat        = Error("unknown format")
)
