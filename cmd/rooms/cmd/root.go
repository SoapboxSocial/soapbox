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

var addr string

func init() {
	rootCmd.AddCommand(server)
	rootCmd.AddCommand(list)
	rootCmd.AddCommand(close)
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
