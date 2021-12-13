package samples_test

import (
	"sort"
	"testing"

	spb "github.com/creachadair/wirepb/samples"

	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

// Explore encoding order for Go generated code.
func TestSamples(t *testing.T) {
	r := &spb.Root{
		Name:    "oak",
		IsAlive: true,
		Branches: []*spb.Branch{
			{Length: 15, LeafColor: spb.Branch_RED},
			{Length: 2, LeafColor: spb.Branch_YELLOW},
			{Length: 2, LeafColor: spb.Branch_GREEN},
			{Length: 9, LeafColor: spb.Branch_BROWN},
		},
	}

	if msg, err := proto.Marshal(r); err != nil {
		t.Fatalf("Marshal: %v", err)
	} else {
		t.Logf("Original wire: %#q", msg)
	}
	t.Log("Original text:\n", prototext.Format(r))

	sort.Slice(r.Branches, func(i, j int) bool {
		if r.Branches[i].Length == r.Branches[j].Length {
			return r.Branches[i].LeafColor < r.Branches[j].LeafColor
		}
		return r.Branches[i].Length < r.Branches[j].Length
	})

	if msg, err := proto.Marshal(r); err != nil {
		t.Fatalf("Marshal: %v", err)
	} else {
		t.Logf("Sorted wire:   %#q", msg)
	}
	t.Log("Sorted text:\n", prototext.Format(r))
}
