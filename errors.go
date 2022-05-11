package ansihtml

import "errors"

var (
	ErrNotSupported   = errors.New("not supported")
	ErrColorUndefined = errors.New("color index is undefined")
	ErrUnexpected     = errors.New("unexpected end")
)
