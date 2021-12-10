package wirepb_test

import (
	"testing"

	"github.com/creachadair/ffs/file/wiretype"
	"github.com/creachadair/wirepb"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

func TestCanonical(t *testing.T) {
	// Wire format ffs.file.Node.
	// https://github.com/creachadair/ffs/blob/default/file/wiretype/wiretype.proto#L24

	const input = "\x12\x13\b\xed\x03\x10\x01\x1a\f\b\xe8\xe9Č\x06\x10" +
		"\xf0\xb3\xfe\x84\x03\")\n\x05audio\x12 \xcb煿<\xae\xe1|\xcd\xdf\bb\xd2" +
		"\xcc\xf2<\xf9$\x85<\x19\xabkpE\x17\xed\x80\x10`S4\")\n\x05blogs\x12 \xd5" +
		"\x14\x1e\xa2\xf66\xa8\xbdu\xaaw\x93\fۇ0\xb9;\xe1\xfc\xa9_\xf0>c\fV\t\x82" +
		"\xa8\x85\x06\"-\n\thousehold\x12 \xfb\x00\xa6\xf0V\x90s\xa7\v~\x13\xa9\xf0" +
		"\x0e\xebE\x05\x8bJ\a\x99\xf5 SN\xb1O4#\xbd\x06\x18\")\n\x05notes\x12 \xdd," +
		"\x84s\xa0\xfd;Y\xe0d\x11\xbbp\xa4\xb4\x16\x9dk\xacYD\x13\x9e\x0f1\x9e\xa8@" +
		"\x99o\x12\xc4\"+\n\arecipes\x12 F\x83\x1e\x8f\x04_Z|\x1e\xc5\x17V5i()\xd2\t" +
		"\x81SZ\x80\xe0\xcd\xdcF\xab {\xe9\x00\x0e\"(\n\x04work\x12 \x15V#̍_\x90\x04V" +
		"\xee\x1fda揁\v.\xdf\xcfM\x0fՊ\x12\xeb\v!\xb26\u007fa\"+\n\awriting\x12 O]" +
		"\x1a3\xd3n\x8d\x98ƪ\xdd%\x1c\xc3z]\xeeD\x0e\xa7\xf4\xd3D5\x86\xda\xfaZ~\xf9N)"

	// Canonicalize the input.
	in := []byte(input)
	var before wiretype.Node
	if err := proto.Unmarshal(in, &before); err != nil {
		t.Fatalf("Unmarshal input: %v", err)
	}
	t.Logf("Input:\n%s", prototext.Format(&before))

	// Decode the output and make sure it survives.
	out := wirepb.Canonical(in)
	var after wiretype.Node
	if err := proto.Unmarshal(out, &after); err != nil {
		t.Fatalf("Unmarshal output: %v", err)
	}
	t.Logf("Output:\n%s", prototext.Format(&after))

	// Verify that the output canonicalizes to itself.
	cmp := wirepb.Canonical(out)
	var canon wiretype.Node
	if err := proto.Unmarshal(cmp, &canon); err != nil {
		t.Fatalf("Unmarshal canon: %v", err)
	}
	t.Logf("Canon:\n%s", prototext.Format(&canon))

	if !proto.Equal(&after, &canon) {
		t.Error("Canonical format not preserved")
	}
}
