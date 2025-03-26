package command

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/schubergphilis/grawsp/internal/meta"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.etcd.io/bbolt"
)

var (
	// Flags
	configFile string

	// Globals
	cache *bbolt.DB

	// Command
	rootCmd = &cobra.Command{
		Use:   meta.AppName,
		Short: "A command line application to manage AWS accounts",
		Long: `A command line application developed to assist engineers manage
their AWS accounts' credentials and resources.`,
		Version: meta.Version,
	}
)

func init() {
	cobra.OnInitialize(onInitialize)
	cobra.OnFinalize(onFinalize)

	rootCmd.Flags().StringVar(&configFile, "config", "", "config file (default is $HOME/.grawsp.yml)")
}

func onFinalize() {
	cache.Close()
}

func onInitialize() {
	// Initialize configuration

	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".grawsp")
	}

	// Defaults
	viper.SetDefault("loglevel", "info")
	viper.SetDefault("orgs", map[string]any{})

	userCacheDir, err := os.UserCacheDir()

	if err != nil {
		panic(err)
	}

	viper.SetDefault("cache_dir", filepath.Join(userCacheDir, meta.AppName))
	viper.AutomaticEnv()

	err = viper.ReadInConfig()

	// Initialize logger

	log.SetOutput(os.Stdout)

	logLevel := strings.ToLower(viper.GetString("loglevel"))

	switch logLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "trace":
		log.SetLevel(log.TraceLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	// Inform if we were able to load the configuration file

	if err == nil {
		log.Debug("Configuration file found: ", viper.ConfigFileUsed())
	} else {
		log.Debug("Configuration file not found")
	}

	// Initialize Cache

	cacheDir := viper.GetString("cache_dir")
	cachePath := filepath.Join(cacheDir, "cache.db")

	log.Debug("Cache location: ", cachePath)

	cache, err = bbolt.Open(cachePath, 0640, nil)

	if err != nil {
		log.Fatal(err)
	}
}

func Execute() error {
	return rootCmd.Execute()
}
