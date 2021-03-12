package cmd

import "github.com/spf13/cobra"

var list = &cobra.Command{
	Use:   "list",
	Short: "list all active rooms",
	RunE:  runList,
}

func runList(*cobra.Command, []string) error {
	return nil
}
