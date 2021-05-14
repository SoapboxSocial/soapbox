package cmd

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "recommendations",
		Short: "Soapbox Recommendations Engine",
		Long:  "",
	}
)

func init() {

}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
