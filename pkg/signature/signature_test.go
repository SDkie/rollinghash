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
	cases := []Case{
		{"../testdata/test1.org", "../testdata/test1.sig"},
		{"../testdata/test2.org", "../testdata/test2.sig"},
		{"../testdata/test3.org", "../testdata/test3.sig"},
		{"../testdata/test4.org", "../testdata/test4.sig"},
		{"../testdata/test5.org", "../testdata/test5.sig"},
		{"../testdata/test6.org", "../testdata/test6.sig"},
		{"../testdata/test7.org", "../testdata/test7.sig"},
		{"../testdata/test8.org", "../testdata/test8.sig"},
		{"../testdata/test9.org", "../testdata/test9.sig"},
		{"../testdata/test10.org", "../testdata/test10.sig"},
		{"../testdata/test11.org", "../testdata/test11.sig"},
		{"../testdata/test12.org", "../testdata/test12.sig"},
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
