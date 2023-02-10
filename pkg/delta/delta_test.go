package delta_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/SDkie/rollinghash/pkg/delta"
	"github.com/SDkie/rollinghash/pkg/util"
	"github.com/google/uuid"
)

type Case struct {
	InputFileName   string
	SigFileName     string
	UpdatedFileName string
	DeltaFileName   string
}

// TestFiles format
// TestX.org    : Original file
// TestX.sig    : Signature file
// TestX.update : Updated file
// TestX.delta  : Delta file

// Infomations about test cases
// Test1 : One Chunk file with no changes
// Test2 : One Chunk file with literals at start
// Test3 : One Chunk file with literals at end

// Test4 : Two Chunk file with no changes
// Test5 : Two Chunk file with literals at start
// Test6 : Two Chunk file with literals at middle
// Test7 : Two Chunk file with literals at end
// Test8 : Two Chunk file with literals at start, middle and end
// Test9 : Two Chunk file with trimmed first chunk
// Test10 : Two Chunk file with some chars replaced in first chunk
// Test11 : Two Chunk file with chunk swapped
// Test12 : Two Chunk file with duplicate chunks in updated file
// Test13 : Two Chunk file with missing first chunk in updated file
// Test14 : Two Chunk file with missing second chunk in updated file
// Test15 : Two Chunk file with updated file having no common data

// Test16 : Small Chunk with some literals at the start
// Test17 : Small Chunk with some literals at the end

// Test18 : Large Chunk with some literals at the start
// Test19 : Large Chunk with some literals at the end

func TestDelta(t *testing.T) {
	var cases []Case
	for i := 1; i <= 19; i++ {
		c := Case{
			InputFileName:   fmt.Sprintf("../testdata/test%d.org", i),
			SigFileName:     fmt.Sprintf("../testdata/test%d.sig", i),
			UpdatedFileName: fmt.Sprintf("../testdata/test%d.update", i),
			DeltaFileName:   fmt.Sprintf("../testdata/test%d.delta", i),
		}
		cases = append(cases, c)
	}

	for i, c := range cases {
		outfile := fmt.Sprintf("%s.sig", uuid.New().String())
		err := delta.GenerateDelta(c.InputFileName, c.SigFileName, c.UpdatedFileName, outfile)
		if err != nil {
			t.Errorf("Test%d failed: error generating delta: %s", i+1, err)
		}

		match, err := util.CompareFileContents(outfile, c.DeltaFileName)
		if err != nil {
			t.Errorf("Test%d failed: error comparing files: %s", i+1, err)
		}
		if !match {
			t.Errorf("Test%d failed: delta mismatch for testfile:%s", i+1, c.InputFileName)
		}

		os.Remove(outfile)
	}
}
