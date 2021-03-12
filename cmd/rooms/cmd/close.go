package cmd

import "github.com/spf13/cobra"

var close = &cobra.Command{
	Use:   "close",
	Short: "close an active room",
	RunE:  runClose,
}

func runClose(*cobra.Command, []string) error {
	return nil
}
