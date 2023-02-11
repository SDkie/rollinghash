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
	// NOTE: check delta_test.go for test case details
	var cases []Case
	for i := 1; i <= 20; i++ {
		c := Case{
			InputFileName:       fmt.Sprintf("../testdata/test%d.org", i),
			ExpectedSigFileName: fmt.Sprintf("../testdata/test%d.sig", i),
		}
		cases = append(cases, c)
	}

	for i, c := range cases {
		outfile := fmt.Sprintf("%s.sig", uuid.New().String())
		_, err := signature.GenerateSignature(c.InputFileName, outfile)
		if err != nil {
			t.Errorf("Test%d failed: error generating signature: %s", i+1, err)
		}

		match, err := util.CompareFileContents(outfile, c.ExpectedSigFileName)
		if err != nil {
			t.Errorf("Test%d failed: error comparing files: %s", i+1, err)
		}
		if !match {
			t.Errorf("Test%d failed: signature mismatch for testfile:%s", i+1, c.InputFileName)
		}

		os.Remove(outfile)
	}
}

func TestEmptyInputFile(t *testing.T) {
	outfile := fmt.Sprintf("%s.sig", uuid.New().String())
	defer os.Remove(outfile)

	_, err := signature.GenerateSignature("../testdata/Test100.org", outfile)
	if err == nil || err != signature.ErrEmptyInputFile {
		t.Fatalf("Test should fail with error:%s", signature.ErrEmptyInputFile)
	}
}
