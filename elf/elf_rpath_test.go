package elf_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/MusicalNinjaDad/snaggle/elf"

	. "github.com/MusicalNinjaDad/snaggle/internal"
)

func TestRPathRunPath(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		shouldError  bool
		errorMessage string
		hasPaths     bool
	}{
		{
			name:         "RUNPATH detected",
			path:         P_rpath_test,
			shouldError:  true,
			errorMessage: "RPATH or RUNPATH",
			hasPaths:     true,
		},
		{
			name:        "no RPATH/RUNPATH",
			path:        P_hello_pie,
			shouldError: false,
			hasPaths:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			Assert := assert.New(t)

			parsed, err := elf.New(tc.path)

			if tc.shouldError {
				Assert.Error(err)
				Assert.ErrorIs(err, elf.ErrRPathOrRunPath)
			} else {
				Assert.NoError(err)
			}

			if tc.hasPaths {
				// When RUNPATH/RPATH is detected, it should be populated
				if len(parsed.Runpath) > 0 || len(parsed.Rpath) > 0 {
					Assert.True(len(parsed.Runpath) > 0 || len(parsed.Rpath) > 0)
				}
			} else {
				// When no RPATH/RUNPATH is present, both should be empty
				Assert.Empty(parsed.Runpath)
				Assert.Empty(parsed.Rpath)
			}
		})
	}
}

func TestRPathRunPathError(t *testing.T) {
	Assert := assert.New(t)

	_, err := elf.New(P_rpath_test)

	// Should have the error in the ErrElf
	var errelf *elf.ErrElf
	Assert.ErrorAs(err, &errelf)
	Assert.ErrorIs(err, elf.ErrRPathOrRunPath)
}
