// Package wirepb implements an experimental canonicalizer for wire format
// protobuf messages.
package wirepb

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

var (
	// ErrBadVarint reports an invalid varint encoding.
	ErrBadVarint = errors.New("invalid varint value")

	// ErrWireType reports an invalid wire type code.
	ErrWireType = errors.New("invalid wire type")

	// ErrShortValue reports a truncated field value.
	ErrShortValue = errors.New("truncated field value")
)

// Constants defining the basic wire types.
const (
	TVarint  = 0
	TFixed64 = 1
	TString  = 2
	TFixed32 = 5
)

// A Scanner parses the lexical structure of a protocol buffer wire message.
// Each successful call to Next advances to the next field of the message, or
// reports an error.
//
// The content of the field can be accessed using the Type, Tag, Field, and
// Value methods. The values reported by these methods are only valid until a
// subsequent call to Next.
type Scanner struct {
	src   []byte
	wtype uint64
	tag   uint64

	pos, end int
	dpos     int
	err      error
}

// NewScanner returns a new scanner that parses the contents of msg.  The
// scanner retains the slice, but does not modify the contents.  The caller
// must also not modify the contents while the scanner is in use, or must make
// a copy before constructing the scanner.
func NewScanner(msg []byte) *Scanner { return &Scanner{src: msg} }

// Next advances the scanner to the next field of the input.
// If no further input is available, Next returns io.EOF.
// Otherwise, errors have concrete type *ScanError.
func (s *Scanner) Next() error {
	s.pos = s.end
	s.wtype, s.tag, s.err = 0, 0, nil

	if s.pos >= len(s.src) {
		s.err = io.EOF
		return s.err
	}
	key, nb := binary.Uvarint(s.src[s.pos:])
	if nb <= 0 {
		return s.fail(s.pos, ErrBadVarint)
	}
	s.dpos = s.pos + nb
	s.wtype = key & 7
	s.tag = key >> 3

	switch s.wtype {
	case TVarint:
		_, nb := binary.Uvarint(s.src[s.dpos:])
		if nb <= 0 {
			return s.fail(s.dpos, ErrBadVarint)
		}
		s.end = s.dpos + nb
		return nil

	case TFixed64:
		if s.dpos+8 > len(s.src) {
			return s.fail(s.dpos, ErrShortValue)
		}
		s.end = s.dpos + 8
		return nil

	case TString:
		slen, nb := binary.Uvarint(s.src[s.dpos:])
		if nb <= 0 {
			return s.fail(s.dpos, ErrBadVarint)
		}
		s.dpos += nb
		end := s.dpos + int(slen)
		if end > len(s.src) {
			return s.fail(s.dpos, ErrShortValue)
		}
		s.end = end
		return nil

	case TFixed32:
		if s.dpos+4 > len(s.src) {
			return s.fail(s.dpos, ErrShortValue)
		}
		s.end = s.dpos + 4
		return nil

	default:
		return s.fail(s.pos, fmt.Errorf("%w: %d", ErrWireType, s.wtype))
	}
}

// Err returns the error reported by the most recent call to Next.
func (s *Scanner) Err() error { return s.err }

// Field returns the slice of the input corresponding to the current field.
// The slice includes the complete tag and value.
func (s *Scanner) Field() []byte { return s.src[s.pos:s.end] }

// Value returns the slice of the input corresponding to the field value.
// The slice includes only the value without the tag or length prefix.
func (s *Scanner) Value() []byte { return s.src[s.dpos:s.end] }

// Type returns the wire type of the current field.
func (s *Scanner) Type() uint64 { return s.wtype }

// Tag returns the tag number of the current field.
func (s *Scanner) Tag() uint64 { return s.tag }

func (s *Scanner) fail(offset int, err error) error {
	s.err = &ScanError{Offset: offset, Err: err}
	return s.err
}

// ScanError is the concrete type of errors reported by the scanner.
type ScanError struct {
	Offset int
	Err    error
}

// Error satisfies the error interface.
func (e *ScanError) Error() string {
	return fmt.Sprintf("offset %d: %v", e.Offset, e.Err)
}

// Unwrap implements error unwrapping for use with errors.Is.
func (e *ScanError) Unwrap() error { return e.Err }
