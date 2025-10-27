// The commandline version of snaggle, for running during container builds etc.
//
//	Snag a copy of FILE and all its dependencies to DESTINATION/bin & DESTINATION/lib64
//
//	Snaggle is designed to help create minimal runtime containers from pre-existing installations.
//	It may work for other use cases and I'd be interested to hear about them at:
//	https://github.com/MusicalNinjaDad/snaggle
//
//	Usage:
//	snaggle FILE DESTINATION [flags]
//
//	Flags:
//	-h, --help   help for snaggle
//
//
//	Snaggle will hardlink (or copy, see notes):
//	- FILE -> DESTINATION/bin
//	- All dynamically linked dependencies -> DESTINATION/lib64
//
//	Note:
//	- Future versions intend to provide improved heuristics for destination paths, currently calling
//	  Snaggle(path/to/a.library.so) will place a.library.so in root/bin and you need to move it manually
//	- Hardlinks will be created if possible.
//	- A copy will be performed if hardlinking fails for one of the following reasons:
//	- path & root are on different filesystems
//	- the user does not have permission to hardlink (e.g.
//	  https://docs.kernel.org/admin-guide/sysctl/fs.html#protected-hardlinks)
//	- Copies will retain the original filemode
//	- Copies will attempt to retain the original ownership, although this will likely fail if running as non-root
//
//	Exit Codes:
//	  0: Success
//	  1: Error
//	  2: Invalid command
//	  3: Panic
package main

import (
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
	rootCmd.SetErrPrefix("snaggle")
	helpTemplate := []string{rootCmd.HelpTemplate(), helpNotes, exitCodes}
	rootCmd.SetHelpTemplate(strings.Join(helpTemplate, "\n"))
	rootCmd.Flags().BoolVar(&inplace, "inplace", false, "Snag in place: only snag dependencies & interpreter")
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
	switch {
	case err == nil:
		os.Exit(0)
	// Safer would be to create a snaggle error and errors.As that for Exit(1)
	case strings.Contains(err.Error(), "accepts"):
		os.Exit(2)
	default:
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "snaggle FILE DESTINATION",
	Short: "Snag a copy of FILE and all its dependencies to DESTINATION/bin & DESTINATION/lib64",
	Long: `Snag a copy of FILE and all its dependencies to DESTINATION/bin & DESTINATION/lib64

Snaggle is designed to help create minimal runtime containers from pre-existing installations.
It may work for other use cases and I'd be interested to hear about them at:
https://github.com/MusicalNinjaDad/snaggle
`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch {
		case inplace:
			return snaggle.Snaggle(args[0], args[1], snaggle.Inplace())
		case recursive:
			return snaggle.Snaggle(args[0], args[1], snaggle.Recursive())
		default:
			return snaggle.Snaggle(args[0], args[1])
		}
	},
}

var helpNotes = `
Snaggle will hardlink (or copy, see notes):
- FILE -> DESTINATION/bin
- All dynamically linked dependencies -> DESTINATION/lib64

Note:
- Future versions intend to provide improved heuristics for destination paths, currently calling
  Snaggle(path/to/a.library.so) will place a.library.so in root/bin and you need to move it manually
- Hardlinks will be created if possible.
- A copy will be performed if hardlinking fails for one of the following reasons:
  - path & root are on different filesystems
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
