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

// Infomations about test cases
// Test1 : One Chunk file with no changes
// Test2 : One Chunk file with literals at start
// Test3 : One Chunk file with literals at end

// Test4 : Two Chunk file with no changes
// Test5 : Two Chunk file with literals at start
// Test6 : Two Chunk file with literals at middle
// Test7 : Two Chunk file with literals at end
// Test8 : Two Chunk file with literals at start, middle and end

func TestDelta(t *testing.T) {
	cases := []Case{
		{"../testdata/test1.org", "../testdata/test1.sig", "../testdata/test1.org", "../testdata/test1.delta"},
		{"../testdata/test2.org", "../testdata/test2.sig", "../testdata/test2.update", "../testdata/test2.delta"},
		{"../testdata/test3.org", "../testdata/test3.sig", "../testdata/test3.update", "../testdata/test3.delta"},
		{"../testdata/test4.org", "../testdata/test4.sig", "../testdata/test4.org", "../testdata/test4.delta"},
		{"../testdata/test5.org", "../testdata/test5.sig", "../testdata/test5.update", "../testdata/test5.delta"},
		{"../testdata/test6.org", "../testdata/test6.sig", "../testdata/test6.update", "../testdata/test6.delta"},
		{"../testdata/test7.org", "../testdata/test7.sig", "../testdata/test7.update", "../testdata/test7.delta"},
		{"../testdata/test8.org", "../testdata/test8.sig", "../testdata/test8.update", "../testdata/test8.delta"},
	}

	for _, c := range cases {
		outfile := fmt.Sprintf("%s.sig", uuid.New().String())
		err := delta.GenerateDelta(c.InputFileName, c.SigFileName, c.UpdatedFileName, outfile)
		if err != nil {
			t.Error("error generating signature: ", err)
		}

		match, err := util.CompareFileContents(outfile, c.DeltaFileName)
		if err != nil {
			t.Error("error comparing files: ", err)
		}
		if !match {
			t.Errorf("signature mismatch for testfile:%s", c.InputFileName)
		}

		os.Remove(outfile)
	}
}
