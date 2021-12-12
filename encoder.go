package wirepb

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
)

// An Encoder is a builder for a wire-format protobuf message.
// A zero value is ready for use.
//
// Each method of the form Name(tag, value) adds a single field to the message
// of the specified type. Use the Message method to construct a field with a
// structured (message) type. By default, fields with zero or empty values are
// omitted from the encoding; set the KeepZeroes option to retain them.
//
// When all the fields have been encoded, use WriteTo to write the encoding
// out. After writing the encoding, the encoder is empty and can be reused for
// another message.
type Encoder struct {
	buf bytes.Buffer

	// If true, emit zero-valued fields explicitly. Otherwise, and by default,
	// zero or empty fields will not be recorded.
	KeepZeroes bool
}

// Bool encodes a Boolean field (varint).
func (e *Encoder) Bool(tag uint64, value bool) {
	if value {
		e.writeField(TVarint, tag, []byte{1})
	} else if e.KeepZeroes {
		e.writeField(TVarint, tag, []byte{0})
	}
}

// Bytes encodes an arbitrary byte array field (string).
func (e *Encoder) Bytes(tag uint64, value []byte) {
	if len(value) != 0 || e.KeepZeroes {
		e.writeStringField(tag, value)
	}
}

// Float32 encodes a floating-point field (fixed32).
func (e *Encoder) Float32(tag uint64, value float32) {
	if value != 0 || e.KeepZeroes {
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], math.Float32bits(value))
		e.writeField(TFixed32, tag, buf[:])
	}
}

// Float64 encodes a floating-point field (fixed64).
func (e *Encoder) Float64(tag uint64, value float64) {
	if value != 0 || e.KeepZeroes {
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], math.Float64bits(value))
		e.writeField(TFixed64, tag, buf[:])
	}
}

// Int64 encodes a signed int64 field (varint).
func (e *Encoder) Int64(tag uint64, value int64) {
	if value != 0 || e.KeepZeroes {
		u := uint64(value<<1) ^ uint64(value>>63)
		e.writeVarintField(tag, u)
	}
}

// Message calls f with an empty encoder for a message field with the given
// tag.  When f returns, the field will be added to the encoder.
func (e *Encoder) Message(tag uint64, f func(*Encoder)) {
	msg := &Encoder{KeepZeroes: e.KeepZeroes}
	f(msg)
	if msg.buf.Len() != 0 || e.KeepZeroes {
		e.writeStringField(tag, msg.buf.Bytes())
	}
}

// Uint64 encodes an unsigned uint64 field (varint).
func (e *Encoder) Uint64(tag, value uint64) {
	if value != 0 || e.KeepZeroes {
		e.writeVarintField(tag, value)
	}
}

// WriteTo writes the current state of the encoding to w.
// This implementation of io.WriterTo never reports an error.
// After a successful call to WriteTo, the encoder is empty and
// can be reused.
func (e *Encoder) WriteTo(w io.Writer) (int64, error) { return e.buf.WriteTo(w) }

// Reset resets the encoder to empty, discarding any data previously written.
func (e *Encoder) Reset() { e.buf.Reset() }

// Encoding returns the current state of the encoded data. The returned slice
// is only valid until the next modification of the encoder.
func (e *Encoder) Encoding() []byte { return e.buf.Bytes() }

func (e *Encoder) writeVarint(v uint64) {
	var buf [10]byte
	n := binary.PutUvarint(buf[:], v)
	e.buf.Write(buf[:n])
}

func (e *Encoder) writeField(wtype, tag uint64, data []byte) {
	e.writeVarint((tag << 3) | wtype)
	e.buf.Write(data)
}

func (e *Encoder) writeStringField(tag uint64, data []byte) {
	e.writeVarint((tag << 3) | TString)
	e.writeVarint(uint64(len(data)))
	e.buf.Write(data)
}

func (e *Encoder) writeVarintField(tag, value uint64) {
	e.writeVarint((tag << 3) | TVarint)
	e.writeVarint(value)
}

// A Message wraps an Encoder to construct a submessage. Add data to the
// message using the ordinary Encoder fields, and call Done to finish the
// message and add to the enclosing field.
type Message struct {
	*Encoder
	tag    uint64
	parent *Encoder
}

// done finishes the message and stores its encoding into the parent field.
func (m *Message) Done() {
	if m.parent == nil {
		panic("message has no parent field")
	}
	m.parent.writeStringField(m.tag, m.buf.Bytes())
	m.parent = nil
}
