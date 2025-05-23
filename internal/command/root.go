package command

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/schubergphilis/grawsp/internal/cacheservice"
	"github.com/schubergphilis/grawsp/internal/meta"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	SessionsBucket = "sessions"
)

var (
	// Flags
	configFile string

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
	cobra.OnInitialize(Initialize)
	cobra.OnFinalize(Finalize)

	rootCmd.Flags().StringVar(&configFile, "config", "", "config file (default is $HOME/.grawsp.yml)")
}

func Finalize() {
	log.Debug("Finalizing...")
	cacheservice.Finalize()
}

func Initialize() {
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
	case "warn":
		log.SetLevel(log.WarnLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	log.Debug("Initializing...")

	// Inform if we were able to load the configuration file

	if err == nil {
		log.Debug("Configuration file found", "file", viper.ConfigFileUsed())
	} else {
		log.Debug("Configuration file not found")
	}

	// Initialize Cache

	cacheDir := viper.GetString("cache_dir")
	cachePath := filepath.Join(cacheDir, "cache.db")

	log.Debug("Using cache", "file", cachePath)

	if err := cacheservice.Init(cachePath); err != nil {
		log.Fatal(err)
	}
}

func Execute() error {
	return rootCmd.Execute()
}
