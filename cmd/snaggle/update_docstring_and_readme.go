//go:build ignore

// Update the docstring in main.go and the codeblock in README.md to reflect the latest help text.
//
// Exit Codes:
//   0: No changes made
//   1: Changes made
//   3: Something more serious went wrong

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"

	. "github.com/MusicalNinjaDad/snaggle/internal"
)

func main() {
	exitcode := 0

	_, thisFile, _, _ := runtime.Caller(0)
	thisDir := filepath.Dir(thisFile)
	workspaceRoot := filepath.Join(thisDir, "../..")

	main_go := filepath.Join(thisDir, "main.go")

	orig_main, err := HashFile(main_go)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(3)
	}

	snaggleBin := Build(nil)
	helptext, err := exec.Command(snaggleBin, "--help").Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(3)
	}

	err = SetDocComment(main_go, string(helptext))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(3)
	}

	err = exec.Command("go", "fmt", main_go).Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(3)
	}

	new_main, err := HashFile(main_go)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(3)
	}

	if !slices.Equal(new_main, orig_main) {
		fmt.Fprintf(os.Stderr, "%s updated", main_go)
		exitcode = 1
	}

	readme_md := filepath.Join(workspaceRoot, "README.md")

	orig_readme, err := HashFile(readme_md)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(3)
	}

	readme_file, err := os.Open(readme_md)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(3)
	}
	defer func() { _ = readme_file.Close() }()

	updatedContents := ReplaceBetween(readme_file, "snaggle --help", "```", helptext)

	readme_file.Close()

	err = os.WriteFile(readme_md, updatedContents, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(3)
	}

	new_readme, err := HashFile(readme_md)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(3)
	}

	if !slices.Equal(new_readme, orig_readme) {
		fmt.Fprintf(os.Stderr, "%s updated", readme_md)
		exitcode = 1
	}

	os.Exit(exitcode)

}
