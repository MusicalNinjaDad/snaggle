package testing

import (
	"iter"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"

	. "github.com/MusicalNinjaDad/snaggle/internal" //lint:ignore ST1001 test helpers
)

type TestCase struct {
	Src            string            // source path
	Dest           string            // destination path
	ExpectedStdout []string          // Split by line, de-(in)dented
	ExpectedFiles  map[string]string // map[original_path]snagged_path
}

// Calls t.Run on the test body for all our test case binaries e.g.:
//
//	for Assert, tc := range TestCases(t) {
//		Assert.NotEmpty(tc)
//	}
func TestCases(t *testing.T) iter.Seq2[*assert.Assertions, TestCase] {
	return func(testbody func(Assert *assert.Assertions, tc TestCase) bool) {
		t.Run("PIE 1 dependency", func(t *testing.T) {
			tc := TestCase{}

			Assert := assert.New(t)

			tc.Src = P_which
			tc.Dest = WorkspaceTempDir(t)
			tc.ExpectedStdout, tc.ExpectedFiles = ExpectedOutput(CommonBinaries(t)["PIE_1"], tc.Dest, false)

			t.Logf("\n\nTestcase details: %s", spew.Sdump(tc))
			testbody(Assert, tc)
		})
	}
}
