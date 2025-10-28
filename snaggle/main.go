/*
The commandline version of snaggle, for running during container builds etc.

Snag a copy of a binary and all its dependencies to DESTINATION/bin & DESTINATION/lib64

Snaggle is designed to help create minimal runtime containers from pre-existing installations.
It may work for other use cases and I'd be interested to hear about them at:
https://github.com/MusicalNinjaDad/snaggle

Usage:

	snaggle [--in-place] FILE DESTINATION
	snaggle [--in-place] [--recursive] DIRECTORY DESTINATION

Flags:

	-h, --help        help for snaggle
	    --in-place    Snag in place: only snag dependencies & interpreter
	-r, --recursive   Recurse subdirectories & snag everything

In the form "snaggle FILE DESTINATION":

	FILE and all dependencies will be snagged to DESTINATION.
	An error will be returned if FILE is not a valid ELF binary.

In the form "snaggle DIRECTORY DESTINATION":

	All valid ELF binaries in DIRECTORY, and all their dependencies, will be snagged to DESTINATION.

Snaggle will hardlink (or copy, see notes):
- Executables              -> DESTINATION/bin
- Dynamic libraries (*.so) -> DESTINATION/lib64

Note:
- Hardlinks will be created if possible.
- A copy will be performed if hardlinking fails for one of the following reasons:
  - FILE/DIRECTORY & DESTINATION are on different filesystems
  - the user does not have permission to hardlink (e.g.
    https://docs.kernel.org/admin-guide/sysctl/fs.html#protected-hardlinks)

- Copies will retain the original filemode
- Copies will attempt to retain the original ownership, although this will likely fail if running as non-root

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

var (
	inplace   bool
	recursive bool
)

func init() {
	log.Default().SetOutput(os.Stdout)
	helpTemplate := []string{rootCmd.HelpTemplate(), helpNotes, exitCodes}
	rootCmd.SetHelpTemplate(strings.Join(helpTemplate, "\n"))
	rootCmd.Flags().BoolVar(&inplace, "in-place", false, "Snag in place: only snag dependencies & interpreter")
	rootCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Recurse subdirectories & snag everything")
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
	err := rootCmd.Execute()
	var snaggleError *snaggle.SnaggleError
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
		var options []snaggle.Option
		if inplace {
			options = append(options, snaggle.InPlace())
		}
		if recursive {
			options = append(options, snaggle.Recursive())
		}
		return snaggle.Snaggle(args[0], args[1], options...)
	},
}

var usages = []string{
	"snaggle [--in-place] FILE DESTINATION",
	"snaggle [--in-place] [--recursive] DIRECTORY DESTINATION",
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

Note:
- Hardlinks will be created if possible.
- A copy will be performed if hardlinking fails for one of the following reasons:
  - FILE/DIRECTORY & DESTINATION are on different filesystems
  - the user does not have permission to hardlink (e.g.
    https://docs.kernel.org/admin-guide/sysctl/fs.html#protected-hardlinks)
- Copies will retain the original filemode
- Copies will attempt to retain the original ownership, although this will likely fail if running as non-root
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
