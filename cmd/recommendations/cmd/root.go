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
		Use:   "recommendations",
		Short: "Soapbox Recommendations Engine",
		Long:  "",
	}
)

type Conf struct {
	Redis   conf.RedisConf    `mapstructure:"redis"`
	DB      conf.PostgresConf `mapstructure:"db"`
	Twitter struct {
		Key    string `mapstructure:"key"`
		Secret string `mapstructure:"secret"`
	} `mapstructure:"twitter"`
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&file, "config", "c", "config.toml", "config file")

	rootCmd.AddCommand(followscmd)
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
