package main

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"

	"github.com/MusicalNinjaDad/snaggle"
)

func init() {
	log.Default().SetOutput(os.Stdout)
}

func main() {
	defer func() {
		if panicking := recover(); panicking != nil {
			fmt.Fprintln(os.Stderr, "Sorry someone panicked!")
			fmt.Fprintln(os.Stderr, "This is what we know ...")
			fmt.Fprintln(os.Stderr, panicking)
			fmt.Fprintln(os.Stderr, string(debug.Stack()))
			os.Exit(3)
		}
	}()
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "snaggle FILE DESTINATION",
	Short: "Snag a copy of FILE and all its dependencies under DESTINATION/bin & DESTINATION/lib64",
	Long: `Snag a copy of FILE and all its dependencies under DESTINATION/bin & DESTINATION/lib64

Snaggle is designed to help create minimal runtime containers from pre-existing installations.
It may work for other use cases and I'd be interested to hear about them at:
https://github.com/MusicalNinjaDad/snaggle

Snaggle will hardlink (or copy, see notes):
- FILE -> DESTINATION/bin
- All dynamically linked dependencies -> DESTINATION/lib64

Note:
Hardlinks will be created if possible.
If for some reason this is not possible, for example FILE & DESTINATION are on different filesystems,
a copy will be performed preserving filemode and attempting to preserve ownership.
`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		panic("foo2")
		return snaggle.Snaggle(args[0], args[1])
	},
}
