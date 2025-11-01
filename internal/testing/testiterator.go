package testing

import (
	"iter"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/MusicalNinjaDad/snaggle/internal" //lint:ignore ST1001 test helpers
)

type TestCase struct {
	Src  string // source path
	Dest string // destination path
}

// Calls t.Run on the test body for all our test case binaries e.g.:
//
//	for Assert, tc := range TestCases(t) {
//		Assert.NotEmpty(tc)
//	}
func TestCases(t *testing.T) iter.Seq2[*assert.Assertions, TestCase] {
	return func(testbody func(Assert *assert.Assertions, tc TestCase) bool) {
		t.Run("PIE 1 dependency", func(t *testing.T) {
			Assert := assert.New(t)
			tc := TestCase{}
			tc.Src = P_which
			tc.Dest = WorkspaceTempDir(t)
			testbody(Assert, tc)
		})
	}
}
