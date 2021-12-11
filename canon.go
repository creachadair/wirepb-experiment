package wirepb

import (
	"bytes"
	"io"
	"sort"
)

// Canonical recursively transforms msg into a canonical representation and
// returns that representation.
//
// If msg is not a valid protobuf wire encoding of a message, Canonical(msg)
// returns a copy of msg, unmodified.
//
// Otherwise, let M be the message encoded by msg, and let marshal(M) denote
// the set of valid protobuf wire encodings of message M (including msg).  For
// any x in marshal(M), Canonical(msg) = Canonical(x).  The canonical string
// for M may not unmarshal to a message equivalent to M, since the order of
// repeated fields may be changed.
//
// Limitations: This implementation requires that the encoding does not contain
// fields of default values. This is ordinarily true for proto3 messages, but
// may not be true for proto2.
func Canonical(msg []byte) []byte {
	buf := make([]byte, len(msg)) // scratch buffer
	cp := make([]byte, len(msg))  // copy of input (permuted in-place)
	copy(cp, msg)
	traverse(buf, cp)
	return cp
}

// traverse recursively rewrites msg into canonical form in-place, using buf as
// temporary working storage.  The contents of buf are garbage after traverse
// returns.
//
// Precondition: len(buf) ≥ len(msg).
func traverse(buf, msg []byte) {
	var fields []entry

	// Attempt to parse the input as a wire-format message.  If this fails,
	// assume it is an arbitrary string and leave it unmodified.
	s := NewScanner(msg)
	for s.Next() == nil {
		fields = append(fields, entry{
			tag:   s.Tag(),
			isStr: s.Type() == TString,
			data:  s.Field(),
			value: s.Value(),
		})
	}
	if s.Err() != io.EOF || len(fields) == 0 {
		return // nothing to do
	}

	// Don't recur until we are sure msg itself is valid.  Otherwise we may
	// permute parts of a non-valid message before discovering the truth.
	for _, e := range fields {
		if e.isStr {
			traverse(buf, e.value)
		}
	}

	// Sort the fields by tag, breaking ties by the lexicographic ordering of
	// their binary contents.
	sort.Sort(entries(fields))
	pos := 0
	for _, e := range fields {
		pos += copy(buf[pos:], e.data)
	}
	if pos != len(msg) {
		panic("invalid message length")
	}
	copy(msg, buf[:pos])
}

type entry struct {
	tag   uint64
	isStr bool
	data  []byte
	value []byte
}

type entries []entry

func (e entries) Len() int      { return len(e) }
func (e entries) Swap(i, j int) { e[i], e[j] = e[j], e[i] }

func (e entries) Less(i, j int) bool {
	if e[i].tag < e[j].tag {
		return true
	} else if e[i].tag > e[j].tag {
		return false
	}
	return bytes.Compare(e[i].value, e[j].value) < 0
}
