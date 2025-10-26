package snaggle

import (
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/ameghdadian/x/iter"
	"github.com/stretchr/testify/assert"

	"github.com/MusicalNinjaDad/snaggle/elf"
	. "github.com/MusicalNinjaDad/snaggle/internal"
	. "github.com/MusicalNinjaDad/snaggle/internal/testing"
)

func TestCommonBinaries3(t *testing.T) {
	tests := CommonBinaries(t)

	for _, tc := range tests {
		t.Run(tc.Description, func(t *testing.T) {
			var stdout strings.Builder
			log.SetOutput(&stdout)
			t.Cleanup(func() { log.SetOutput(os.Stdout) })
			i, _ := strconv.Atoi(log.Prefix())
			log.SetPrefix(strconv.Itoa(i + 1))

			Assert := assert.New(t)
			tmp := WorkspaceTempDir(t)

			binPath := filepath.Join(tmp, "bin", filepath.Base(tc.Elf.Name))
			expectedOut := make([]string, 0, 1+len(tc.Elf.Dependencies))
			expectedOut = append(expectedOut, tc.Elf.Path+" -> "+binPath)
			// TODO: #51 ugly - should be in the tc - needs a tidy
			if tc.Elf.Type == elf.PIE {
				expectedOut = append(expectedOut, tc.Elf.Interpreter+" -> "+filepath.Join(tmp, P_ld_linux))
			}
			var libCopies []string
			for _, lib := range tc.Elf.Dependencies {
				copy := filepath.Join(tmp, "lib64", filepath.Base(lib))
				libCopies = append(libCopies, copy)
				expectedOut = append(expectedOut, lib+" -> "+copy)
			}

			err := Snaggle(tc.Elf.Path, tmp)
			Assert.NoError(err)
			AssertSameInode(t, tc.Elf.Path, binPath)
			for idx, copy := range libCopies {
				original := tc.Elf.Dependencies[idx]
				same := SameFile(original, copy)
				assert.Truef(t, same, "%s & %s are different files", original, copy)
			}

			Assert.ElementsMatch(expectedOut, slices.Collect(iter.Map((strings.Lines(stdout.String())), strings.TrimSpace)))
		})
	}
}
