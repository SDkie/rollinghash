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
		name     string
		testNo   int
		expError error
	}{
		// Happy Paths
		{name: "One Chunk file with no changes", testNo: 1, expError: nil},
		{name: "One Chunk file with literals at start", testNo: 2, expError: nil},
		{name: "One Chunk file with literals at end", testNo: 3, expError: nil},

		{name: "Two Chunk file with no changes", testNo: 4, expError: nil},
		{name: "Two Chunk file with literals at start", testNo: 5, expError: nil},
		{name: "Two Chunk file with literals at middle", testNo: 6, expError: nil},
		{name: "Two Chunk file with literals at end", testNo: 7, expError: nil},
		{name: "Two Chunk file with literals at start, middle and end", testNo: 8, expError: nil},
		{name: "Two Chunk file with trimmed first chunk", testNo: 9, expError: nil},
		{name: "Two Chunk file with some chars replaced in first chunk", testNo: 10, expError: nil},
		{name: "Two Chunk file with chunk swapped", testNo: 11, expError: nil},
		{name: "Two Chunk file with duplicate chunks in updated file", testNo: 12, expError: nil},
		{name: "Two Chunk file with missing first chunk in updated file", testNo: 13, expError: nil},
		{name: "Two Chunk file with missing second chunk in updated file", testNo: 14, expError: nil},
		{name: "Two Chunk file with updated file having no common data", testNo: 15, expError: nil},

		{name: "Small Chunk with some literals at the start", testNo: 16, expError: nil},
		{name: "Small Chunk with some literals at the end", testNo: 17, expError: nil},

		{name: "Large Chunk with some literals at the start", testNo: 18, expError: nil},
		{name: "Large Chunk with some literals at the end", testNo: 19, expError: nil},
		{name: "Large Chunk with some literals missing in the middle", testNo: 20, expError: nil},

		// Unhappy Paths
		{name: "Empty Original file", testNo: 101, expError: delta.ErrEmptyOriginalFile},
		{name: "Empty Updated file", testNo: 102, expError: delta.ErrEmptyUpdatedFile},
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
			if err != c.expError {
				t.Fatalf("'%s' Failed : expected error:%v, got:%v", t.Name(), c.expError, err)
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
