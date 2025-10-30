//go:build ignore

// Update the docstring in main.go and the codeblock in README.md to reflect the latest help text.
//
// Exit Codes:
//   0: No changes made
//   1: Changes made
//   3: Something more serious went wrong

package main

import (
	"os"
	"os/exec"
	"slices"

	. "github.com/MusicalNinjaDad/snaggle/internal"
)

func main() {
	exitcode := 0
	main_go, err := HashFile("main.go")
	if err != nil {
		println(err)
		os.Exit(3)
	}

	snaggleBin := Build(nil)
	helptext, err := exec.Command(snaggleBin, "--help").Output()
	if err != nil {
		println(err)
		os.Exit(3)
	}

	err = SetDocComment("main.go", string(helptext))
	if err != nil {
		println(err)
		os.Exit(3)
	}

	new_main, err := HashFile("main.go")
	if err != nil {
		println(err)
		os.Exit(3)
	}

	if !slices.Equal(new_main, main_go) {
		println("main.go updated")
		exitcode = 1
	}

	os.Exit(exitcode)

}
