// Helpers and values which can be used in any package
package internal

import (
	"slices"
	"strings"
	"testing"

	"github.com/ameghdadian/x/iter"

	"github.com/stretchr/testify/assert"
)

// Test Helper: validates that a slice dependency paths are equal without breaking cross-platform
// portability (libs are potentially in different paths)
func AssertDependenciesEqual(t *testing.T, expected []string, actual []string) {
	t.Helper()
	if len(expected) == 0 && len(actual) == 0 {
		return // no dependencies
	}
	for idx, dep := range expected {
		assert.Zerof(t, Libpathcmp(dep, actual[idx]), "dependency %v differs: %s != %s", idx, dep, actual[idx])
	}
}

// StripLines converts a multiline string into a []string without line termination or indentation.
func StripLines(multiline string) []string {
	return slices.Collect(iter.Map(strings.Lines(multiline), strings.TrimSpace))
}

// Build a message about a link attempt for stdout
func LinkMessage(from string, to string) string {
	return from + " -> " + to
}
