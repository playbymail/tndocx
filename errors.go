// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package tndocx

type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrEmptyInput     = Error("empty input")
	ErrNotImplemented = Error("not implemented")
	ErrUnknownFormat  = Error("unknown format")
)
