package wirepb_test

import (
	"io"
	"testing"

	"github.com/creachadair/wirepb"
)

func TestEncoder(t *testing.T) {
	enc := &wirepb.Encoder{KeepZeroes: true}
	enc.Message(1, func(e *wirepb.Encoder) {
		e.Uint64(3, 0)
		e.Bytes(5, []byte("hello"))
		e.Int64(9, 100)
	})
	enc.Message(2, func(e *wirepb.Encoder) {
		e.Float32(1, 3.1415)
		e.Bool(2, true)
		e.Bool(2, false)
		e.Bool(2, true)
	})
	enc.Message(3, func(e *wirepb.Encoder) {})
	t.Logf("Encoded: %#q", enc.Encoding())

	type entry struct {
		wtype, tag uint64
		value      []byte
	}

	var visit func(string, []byte)
	visit = func(pfx string, msg []byte) {
		var fields []entry
		s := wirepb.NewScanner(msg)
		for s.Next() == nil {
			fields = append(fields, entry{s.Type(), s.Tag(), s.Value()})
		}
		if s.Err() != io.EOF {
			return
		}
		for _, field := range fields {
			t.Logf("%sField: type=%d tag=%d value=%#q", pfx, field.wtype, field.tag, field.value)
			if field.wtype == wirepb.TString {
				visit(pfx+"  ", field.value)
			}
		}
	}
	visit("", enc.Encoding())
}
