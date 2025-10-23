//go:build testpanic

package main

import "github.com/spf13/cobra"

func init() {
	rootCmd.PreRun = func(cmd *cobra.Command, args []string) {
		panic("you got a special testing build that always panics. (Tip: don't build with `-tags testpanic`)")
	}
}
