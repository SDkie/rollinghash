package delta_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/SDkie/rollinghash/pkg/delta"
	"github.com/SDkie/rollinghash/pkg/signature"
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
// Test20 : Large Chunk with some literals missing in the middle

// Error Cases:
// Test100 : Empty input file for signature
// Test101 : Empty original file for delta
// Test102 : Invalid signature file for delta
// Test103 : Invalid chunk size
// Test104 : Empty updated file

func TestDelta(t *testing.T) {
	var cases []Case
	for i := 1; i <= 20; i++ {
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

func TestEmptyOriginalFile(t *testing.T) {
	outfile := fmt.Sprintf("%s.sig", uuid.New().String())
	defer os.Remove(outfile)

	err := delta.GenerateDelta("../testdata/Test101.org", "../testdata/Test101.sig", "../testdata/Test101.update", outfile)
	if err == nil {
		t.Fatalf("Test should fail with error:%s", delta.ErrEmptyOriginalFile)
	}

	if err != delta.ErrEmptyOriginalFile {
		t.Fatalf("Test should fail with error:%s", delta.ErrEmptyOriginalFile)
	}
}

func TestInvalidSignatureFile(t *testing.T) {
	outfile := fmt.Sprintf("%s.sig", uuid.New().String())
	defer os.Remove(outfile)

	err := delta.GenerateDelta("../testdata/Test102.org", "../testdata/Test102.sig", "../testdata/Test102.update", outfile)
	if err == nil {
		t.Fatalf("Test should fail with error:%s", signature.ErrInvalidSignatureFile)
	}

	if err != signature.ErrInvalidSignatureFile {
		t.Fatalf("Test should fail with error:%s", signature.ErrInvalidSignatureFile)
	}
}

func TestInvalidChunkSize(t *testing.T) {
	outfile := fmt.Sprintf("%s.sig", uuid.New().String())
	defer os.Remove(outfile)

	err := delta.GenerateDelta("../testdata/Test103.org", "../testdata/Test103.sig", "../testdata/Test103.update", outfile)
	if err == nil {
		t.Fatalf("Test should fail with error:%s", signature.ErrInvalidChunkSize)
	}

	if err != signature.ErrInvalidChunkSize {
		t.Fatalf("Test should fail with error:%s", signature.ErrInvalidChunkSize)
	}
}

func TestEmptyUpdatedFile(t *testing.T) {
	outfile := fmt.Sprintf("%s.sig", uuid.New().String())
	defer os.Remove(outfile)

	err := delta.GenerateDelta("../testdata/Test104.org", "../testdata/Test104.sig", "../testdata/Test104.update", outfile)
	if err == nil {
		t.Fatalf("Test should fail with error:%s", delta.ErrEmptyUpdatedFile)
	}

	if err != delta.ErrEmptyUpdatedFile {
		t.Fatalf("Test should fail with error:%s", delta.ErrEmptyUpdatedFile)
	}
}
