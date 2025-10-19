package test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/MusicalNinjaDad/snaggle/internal"
)

func Pwd(t *testing.T) string {
	t.Helper()
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal("Failed to get pwd. Error:", err)
	}
	return pwd
}

func AssertDependenciesEqual(t *testing.T, expected []string, actual []string) {
	t.Helper()
	for idx, dep := range expected {
		assert.Zero(t, internal.Libpathcmp(dep, actual[idx]))
	}
}
