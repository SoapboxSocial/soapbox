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
	list.PersistentFlags().StringVarP(&addr, "addr", "a", "127:0:0:1:50052", "grpc address")

	rootCmd.AddCommand(server)
	rootCmd.AddCommand(list)
	rootCmd.AddCommand(close)
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
