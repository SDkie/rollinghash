package delta_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/SDkie/rollinghash/pkg/delta"
	"github.com/SDkie/rollinghash/pkg/util"
	"github.com/google/uuid"
)

// TestFiles format
// TestX.org    : Original file
// TestX.sig    : Signature file
// TestX.update : Updated file
// TestX.delta  : Delta file

func TestGenerateDelta(t *testing.T) {
	cases := []struct {
		name   string
		testNo int
		err    error
	}{
		// Happy Paths
		{name: "One Chunk file with no changes", testNo: 1, err: nil},
		{name: "One Chunk file with literals at start", testNo: 2, err: nil},
		{name: "One Chunk file with literals at end", testNo: 3, err: nil},

		{name: "Two Chunk file with no changes", testNo: 4, err: nil},
		{name: "Two Chunk file with literals at start", testNo: 5, err: nil},
		{name: "Two Chunk file with literals at middle", testNo: 6, err: nil},
		{name: "Two Chunk file with literals at end", testNo: 7, err: nil},
		{name: "Two Chunk file with literals at start, middle and end", testNo: 8, err: nil},
		{name: "Two Chunk file with trimmed first chunk", testNo: 9, err: nil},
		{name: "Two Chunk file with some chars replaced in first chunk", testNo: 10, err: nil},
		{name: "Two Chunk file with chunk swapped", testNo: 11, err: nil},
		{name: "Two Chunk file with duplicate chunks in updated file", testNo: 12, err: nil},
		{name: "Two Chunk file with missing first chunk in updated file", testNo: 13, err: nil},
		{name: "Two Chunk file with missing second chunk in updated file", testNo: 14, err: nil},
		{name: "Two Chunk file with updated file having no common data", testNo: 15, err: nil},

		{name: "Small Chunk with some literals at the start", testNo: 16, err: nil},
		{name: "Small Chunk with some literals at the end", testNo: 17, err: nil},

		{name: "Large Chunk with some literals at the start", testNo: 18, err: nil},
		{name: "Large Chunk with some literals at the end", testNo: 19, err: nil},
		{name: "Large Chunk with some literals missing in the middle", testNo: 20},

		// Unhappy Paths
		{name: "Empty Original file", testNo: 101, err: delta.ErrEmptyOriginalFile},
		{name: "Empty Updated file", testNo: 102, err: delta.ErrEmptyUpdatedFile},
	}

	for _, c := range cases {
		tf := func(t *testing.T) {
			inputfile := fmt.Sprintf("testdata/test%d.org", c.testNo)
			sigfile := fmt.Sprintf("testdata/test%d.sig", c.testNo)
			updatedfile := fmt.Sprintf("testdata/test%d.update", c.testNo)
			expectedDeltafile := fmt.Sprintf("testdata/test%d.delta", c.testNo)

			deltafile := fmt.Sprintf("testdata/%s.delta", uuid.New().String())
			defer os.Remove(deltafile)

			err := delta.GenerateDelta(inputfile, sigfile, updatedfile, deltafile)
			if err != c.err {
				t.Fatalf("'%s' Failed : expected error:%v, got:%v", t.Name(), c.err, err)
			}
			if err != nil {
				return
			}

			match, err := util.CompareFileContents(deltafile, expectedDeltafile)
			if err != nil {
				t.Fatalf("'%s' Failed with error: %v", t.Name(), err)
			}
			if !match {
				t.Fatalf("'%s' Failed : delta file contents do not match", t.Name())
			}
		}

		t.Run(c.name, tf)
	}
}
