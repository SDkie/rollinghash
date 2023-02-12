package signature_test

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/SDkie/rollinghash/pkg/signature"
	"github.com/SDkie/rollinghash/pkg/util"
	"github.com/google/uuid"
)

// TestFiles format
// TestX.org : Input file
// TestX.sig : Signature file

func TestGenerateSignature(t *testing.T) {
	cases := []struct {
		name     string
		testNo   int
		expError error
	}{
		// Happy Paths
		{name: "One Chunk file", testNo: 1, expError: nil},
		{name: "Two Chunk file", testNo: 2, expError: nil},
		{name: "Three Chunk file", testNo: 3, expError: nil},
		{name: "Small Chunk file", testNo: 4, expError: nil},
		{name: "Big Chunk file", testNo: 5, expError: nil},

		// Unhappy Paths
		{name: "Empty Input file", testNo: 101, expError: signature.ErrEmptyInputFile},
	}

	for _, c := range cases {
		tf := func(t *testing.T) {
			inputfile := fmt.Sprintf("testdata/test%d.org", c.testNo)
			expectedSigfile := fmt.Sprintf("testdata/test%d.sig", c.testNo)

			sigfile := fmt.Sprintf("testdata/%s.delta", uuid.New().String())
			defer os.Remove(sigfile)

			_, err := signature.GenerateSignature(inputfile, sigfile)
			if err != c.expError {
				t.Fatalf("'%s' Failed : expected error:%v, got:%v", t.Name(), c.expError, err)
			}
			if err != nil {
				return
			}

			match, err := util.CompareFileContents(sigfile, expectedSigfile)
			if err != nil {
				t.Fatalf("'%s' Failed with error : %v", t.Name(), err)
			}
			if !match {
				t.Fatalf("'%s' Failed : signature file contents do not match", t.Name())
			}
		}

		t.Run(c.name, tf)
	}
}

func TestReadSignature(t *testing.T) {
	cases := []struct {
		name         string
		testNo       int
		expSignature signature.Signature
		expError     error
	}{
		// Happy Paths
		{name: "One Chunk file", testNo: 1, expSignature: signature.Signature{ChunkLen: 256, TotalChunks: 1, Hashes: []uint32{3963550426}}, expError: nil},
		{name: "Two Chunk file", testNo: 2, expSignature: signature.Signature{ChunkLen: 256, TotalChunks: 2, Hashes: []uint32{3963550426, 1999309273}}, expError: nil},
		{name: "Three Chunk file", testNo: 3, expSignature: signature.Signature{ChunkLen: 256, TotalChunks: 3, Hashes: []uint32{3963550426, 1999309273, 35068120}}, expError: nil},
		{name: "Small Chunk file", testNo: 4, expSignature: signature.Signature{ChunkLen: 256, TotalChunks: 1, Hashes: []uint32{4150264061}}, expError: nil},

		// Unhappy Paths
		{name: "Invalid signature file", testNo: 102, expError: signature.ErrInvalidSignatureFile},
		{name: "Invalid chunk size", testNo: 103, expError: signature.ErrInvalidChunkSize},
	}

	for _, c := range cases {
		tf := func(t *testing.T) {
			sigfile := fmt.Sprintf("testdata/test%d.sig", c.testNo)
			signature, err := signature.ReadSignature(sigfile)
			if err != c.expError {
				t.Fatalf("'%s' Failed : expected error:%v, got:%v", t.Name(), c.expError, err)
			}
			if err != nil {
				return
			}

			if !reflect.DeepEqual(*signature, c.expSignature) {
				t.Fatalf("'%s' Failed : signature does not match", t.Name())
			}
		}

		t.Run(c.name, tf)
	}
}
