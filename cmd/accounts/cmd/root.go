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
		Use:   "accounts",
		Short: "Soapbox Third-Party Accounts Management",
		Long:  "",
	}
)

type Conf struct {
	DB      conf.PostgresConf `mapstructure:"db"`
	Twitter struct {
		Key    string `mapstructure:"key"`
		Secret string `mapstructure:"secret"`
	} `mapstructure:"twitter"`
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&file, "config", "c", "config.toml", "config file")

	rootCmd.AddCommand(twitterCmd)
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
