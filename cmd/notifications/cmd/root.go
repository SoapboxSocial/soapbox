package cmd

import "github.com/spf13/cobra"

var (
	rootCmd = &cobra.Command{
		Use:   "notifications",
		Short: "Soapbox Notifications",
		Long:  "",
	}
)

func init() {
	rootCmd.AddCommand(workerCmd)
	rootCmd.AddCommand(send)
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
