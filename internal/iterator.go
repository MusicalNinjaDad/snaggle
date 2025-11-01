package internal

import (
	"iter"
	"testing"
)

func tcs(t *testing.T) iter.Seq[string] {
	return func(testbody func(string) bool) {
		t.Run("foo", func(t *testing.T) {
			testbody("foo")
		})
		t.Run("bar", func(t *testing.T) {
			testbody("bar")
		})
	}
}
