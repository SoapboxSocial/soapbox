package cmd

import "github.com/spf13/cobra"

var follows = &cobra.Command{
	Use:   "follows",
	Short: "created recommendations for users to follow",
	RunE:  runFollows,
}

func runFollows(*cobra.Command, []string) error {
	return nil
}
