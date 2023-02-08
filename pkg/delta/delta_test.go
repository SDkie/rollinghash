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

func TestDelta(t *testing.T) {
	cases := []Case{
		{"../testdata/test1.org", "../testdata/test1.sig", "../testdata/test1.org", "../testdata/test1.delta"},
		{"../testdata/test2.org", "../testdata/test2.sig", "../testdata/test2.update", "../testdata/test2.delta"},
		{"../testdata/test3.org", "../testdata/test3.sig", "../testdata/test3.update", "../testdata/test3.delta"},
		{"../testdata/test4.org", "../testdata/test4.sig", "../testdata/test4.org", "../testdata/test4.delta"},
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
