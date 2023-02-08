package signature_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/SDkie/rollinghash/pkg/signature"
	"github.com/SDkie/rollinghash/pkg/util"
	"github.com/google/uuid"
)

type Case struct {
	InputFileName       string
	ExpectedSigFileName string
}

func TestSignature(t *testing.T) {
	var cases []Case

	for i := 1; i <= 14; i++ {
		c := Case{
			InputFileName:       fmt.Sprintf("../testdata/test%d.org", i),
			ExpectedSigFileName: fmt.Sprintf("../testdata/test%d.sig", i),
		}
		cases = append(cases, c)
	}

	for _, c := range cases {
		outfile := fmt.Sprintf("%s.sig", uuid.New().String())
		err := signature.GenerateSignature(c.InputFileName, outfile)
		if err != nil {
			t.Error("error generating signature: ", err)
		}

		match, err := util.CompareFileContents(outfile, c.ExpectedSigFileName)
		if err != nil {
			t.Error("error comparing files: ", err)
		}
		if !match {
			t.Errorf("signature mismatch for testfile:%s", c.InputFileName)
		}

		os.Remove(outfile)
	}
}
