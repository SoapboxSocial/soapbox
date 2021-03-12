package cmd

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "rooms",
		Short: "Soapbox Room Servers",
		Long:  "",
	}
)

func init() {
	rootCmd.AddCommand(server)
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
