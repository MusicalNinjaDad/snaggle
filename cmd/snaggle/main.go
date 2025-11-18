//go:generate go run update_docstring_and_readme.go

/*
Snag a copy of a binary and all its dependencies to DESTINATION/bin & DESTINATION/lib64

Snaggle is designed to help create minimal runtime containers from pre-existing installations.
It may work for other use cases and I'd be interested to hear about them at:
https://github.com/MusicalNinjaDad/snaggle

Usage:

	snaggle [--in-place] FILE DESTINATION
	snaggle [--copy | --in-place] [--recursive] DIRECTORY DESTINATION

Flags:

	    --copy        Copy entire directory contents to /DESTINATION/full/source/path
	-h, --help        help for snaggle
	    --in-place    Snag in place: only snag dependencies & interpreter
	-r, --recursive   Recurse subdirectories & snag everything
	-v, --verbose     Output to stdout and process sequentially for readability
	    --version     version for snaggle

In the form "snaggle FILE DESTINATION":

	FILE and all dependencies will be snagged to DESTINATION.
	An error will be returned if FILE is not a valid ELF binary.

In the form "snaggle DIRECTORY DESTINATION":

	All valid ELF binaries in DIRECTORY, and all their dependencies, will be snagged to DESTINATION.

Snaggle will hardlink (or copy, see notes):
- Executables              -> DESTINATION/bin
- Dynamic libraries (*.so) -> DESTINATION/lib64

Notes:
  - Follows symlinks
  - Hardlinks will be created if possible.
  - A copy will be performed if hardlinking fails for one of the following reasons:
    FILE/DIRECTORY & DESTINATION are on different filesystems or
    the user does not have permission to hardlink (e.g.
    https://docs.kernel.org/admin-guide/sysctl/fs.html#protected-hardlinks)
  - Copies will retain the original filemode
  - Copies will attempt to retain the original ownership, although this will likely fail if running as non-root
  - Running with --verbose will be slower, not only due to processing stdout, but also as each file will be processed
    sequentially to provide readable output. Running silently will process all files and dependencies in parallel.

Exit Codes:

	0: Success
	1: Error
	2: Invalid command
	3: Panic
*/
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"

	"github.com/MusicalNinjaDad/snaggle"
)

var options []snaggle.Option

func addOption(option snaggle.Option) func(string) error {
	return func(_ string) error {
		options = append(options, option)
		return nil
	}
}

func init() {
	log.Default().SetOutput(os.Stdout)

	rootCmd.Version = snaggle.Version

	helpTemplate := []string{rootCmd.HelpTemplate(), helpNotes, exitCodes}
	rootCmd.SetHelpTemplate(strings.Join(helpTemplate, "\n"))

	rootCmd.Flags().BoolFunc("copy", "Copy entire directory contents to /DESTINATION/full/source/path", addOption(snaggle.Copy()))
	rootCmd.Flags().BoolFunc("in-place", "Snag in place: only snag dependencies & interpreter", addOption(snaggle.InPlace()))
	rootCmd.Flags().BoolFuncP("recursive", "r", "Recurse subdirectories & snag everything", addOption(snaggle.Recursive()))
	rootCmd.Flags().BoolFuncP("verbose", "v", "Output to stdout and process sequentially for readability", addOption(snaggle.Verbose()))

	// These are called somewhere in execute - which is not available to integration tests
	rootCmd.InitDefaultHelpFlag()
	rootCmd.InitDefaultVersionFlag()
}

// defer panicHandler to get meaningful output to stderr and control over the exitcode on panic
//
// panicHandler calls os.Exit(exitcode) - so defer it as early as possible, any remaining functions
// in the defer queue will be skipped
func panicHandler(exitcode int) {
	if panicking := recover(); panicking != nil {
		fmt.Fprintln(os.Stderr, "Sorry someone panicked!")
		fmt.Fprintln(os.Stderr, "This is what we know ...")
		fmt.Fprintln(os.Stderr, panicking)
		fmt.Fprintln(os.Stderr, string(debug.Stack()))
		os.Exit(exitcode)
	}
}

func main() {
	defer panicHandler(3)

	var snaggleError *snaggle.SnaggleError
	err := rootCmd.Execute()
	switch {
	case err == nil:
		os.Exit(0)
	case errors.As(err, &snaggleError):
		os.Exit(1)
	default:
		println(rootCmd.UsageString())
		os.Exit(2)
	}
}

var rootCmd = &cobra.Command{
	Use:                   strings.Join(usages, "\n  "),
	SilenceUsage:          true,
	DisableFlagsInUseLine: true,
	Long: `Snag a copy of a binary and all its dependencies to DESTINATION/bin & DESTINATION/lib64

Snaggle is designed to help create minimal runtime containers from pre-existing installations.
It may work for other use cases and I'd be interested to hear about them at:
https://github.com/MusicalNinjaDad/snaggle
`,
	Args: ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return snaggle.Snaggle(args[0], args[1], options...)
	},
}

var usages = []string{
	"snaggle [--in-place] FILE DESTINATION",
	"snaggle [--copy | --in-place] [--recursive] DIRECTORY DESTINATION",
}

var helpNotes = `
In the form "snaggle FILE DESTINATION":
  FILE and all dependencies will be snagged to DESTINATION.
  An error will be returned if FILE is not a valid ELF binary.

In the form "snaggle DIRECTORY DESTINATION":
  All valid ELF binaries in DIRECTORY, and all their dependencies, will be snagged to DESTINATION.

Snaggle will hardlink (or copy, see notes):
- Executables              -> DESTINATION/bin
- Dynamic libraries (*.so) -> DESTINATION/lib64

Notes:
- Follows symlinks
- Hardlinks will be created if possible.
- A copy will be performed if hardlinking fails for one of the following reasons:
    FILE/DIRECTORY & DESTINATION are on different filesystems or
    the user does not have permission to hardlink (e.g.
      https://docs.kernel.org/admin-guide/sysctl/fs.html#protected-hardlinks)
- Copies will retain the original filemode
- Copies will attempt to retain the original ownership, although this will likely fail if running as non-root
- Running with --verbose will be slower, not only due to processing stdout, but also as each file will be processed
  sequentially to provide readable output. Running silently will process all files and dependencies in parallel.
`

var exitCodes = `Exit Codes:
  0: Success
  1: Error
  2: Invalid command
  3: Panic
`

func ExactArgs(n int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != n {
			return fmt.Errorf("snaggle expects %d argument(s), %d received", n, len(args))
		}
		return nil
	}
}
