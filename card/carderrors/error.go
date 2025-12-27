// Package carderrors defines error types used when working with smart cards.
package carderrors

import "errors"

// ErrInvalidLength is returned when data has invalid length.
var ErrInvalidLength = errors.New("invalid length")

// ErrInvalidFormat is returned when data has invalid format.
var ErrInvalidFormat = errors.New("invalid format")
