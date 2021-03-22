package cmd

import (
	"log"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/spf13/cobra"

	"github.com/soapboxsocial/soapbox/pkg/conf"
	"github.com/soapboxsocial/soapbox/pkg/groups"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

var (
	file   string
	config *Conf

	// Used by some of the commands.
	client        *elasticsearch.Client
	userBackend   *users.UserBackend
	groupsBackend *groups.Backend

	rootCmd = &cobra.Command{
		Use:   "indexer",
		Short: "Soapbox Search Indexing",
		Long:  "",
	}
)

type Conf struct {
	Redis conf.RedisConf    `mapstructure:"redis"`
	DB    conf.PostgresConf `mapstructure:"db"`
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&file, "config", "c", "config.toml", "config file")
	rootCmd.AddCommand(worker)
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
