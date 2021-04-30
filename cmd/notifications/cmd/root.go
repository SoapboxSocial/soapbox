package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/soapboxsocial/soapbox/pkg/conf"
)

var (
	file   string
	config *Conf

	rootCmd = &cobra.Command{
		Use:   "notifications",
		Short: "Soapbox Notifications",
		Long:  "",
	}
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&file, "config", "c", "config.toml", "config file")

	rootCmd.AddCommand(workerCmd)
	rootCmd.AddCommand(send)
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func initConfig() {
	config = &Conf{}
	err := conf.Load(file, config)
	if err != nil {
		log.Fatalf("failed to load config: %s", err)
	}
}
