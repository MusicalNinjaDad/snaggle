// Helpers and values which can be used in any package
package internal

import (
	"slices"
	"strings"

	"github.com/ameghdadian/x/iter"
)

// StripLines converts a multiline string into a []string without line termination or indentation.
func StripLines(multiline string) []string {
	return slices.Collect(iter.Map(strings.Lines(multiline), strings.TrimSpace))
}

// Build a message about a link attempt for stdout
func LinkMessage(from string, to string) string {
	return from + " -> " + to
}
