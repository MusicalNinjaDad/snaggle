package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIt(t *testing.T) {
	for range tcs(t) {
		assert.True(t, true)
	}
}
