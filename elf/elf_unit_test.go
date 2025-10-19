package elf

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func pwd(t *testing.T) string {
	t.Helper()
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal("Failed to get pwd. Error:", err)
	}
	return pwd
}

func TestLdd(t *testing.T) {
	Assert := assert.New(t)
	which := filepath.Join(pwd(t), "../testdata/which")
	expectedDependencies := []string{"/lib64/libc.so.6"}
	dependencies, err := ldd(which)
	Assert.NoError(err)
	Assert.ElementsMatch(expectedDependencies, dependencies)
}
