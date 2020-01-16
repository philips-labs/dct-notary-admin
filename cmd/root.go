package cmd

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	homedir "github.com/mitchellh/go-homedir"

	"github.com/philips-labs/dct-notary-admin/lib"
	"github.com/philips-labs/dct-notary-admin/lib/notary"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "dctna",
	Short: "Docker Content Trust and Notary Admin",
	Long: `Docker Content Trust and Notary Admin allows to
create new targets and manage signers / delegates via a
RESTFULL api. This enables us to keep keys more private and
centralized to better manage backups.`,
	Run: func(cmd *cobra.Command, args []string) {
		if v, _ := cmd.Flags().GetBool("version"); v {
			cmd.Println(sprintVersion(cmd))
			return
		}

		logger, err := zap.NewDevelopment(zap.AddStacktrace(zapcore.FatalLevel))
		if err != nil {
			log.Fatalf("Can't initialize zap logger: %v", err)
		}
		defer logger.Sync()

		serverCfg, err := unmarshalServerConfig()
		if err != nil {
			logger.Fatal("Could not parse configuration", zap.Error(err))
		}
		logger.Debug("Unmarshalled ServerConfig", zap.Any("config", serverCfg))

		notaryCfg, err := unmarshalNotaryConfig()
		if err != nil {
			logger.Fatal("Could not parse configuration", zap.Error(err))
		}
		logger.Debug("Unmarshalled NotaryConfig", zap.Any("config", notaryCfg))

		n := notary.NewService(notaryCfg)
		server := lib.NewServer(serverCfg, n, logger)
		server.Start()
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./.notary/config.json or $HOME/.notary/config.json)")
	rootCmd.PersistentFlags().String("listen-addr", "", "http listen address of server")
	rootCmd.PersistentFlags().String("listen-addr-tls", "", "https listen address of server")

	rootCmd.Flags().BoolP("version", "v", false, "shows version information")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(path.Join(home, ".notary"))
		viper.AddConfigPath(path.Join("./", ".notary"))
		viper.SetConfigName("config")
	}

	setDefaultAndFlagBinding("server.listen_addr", "listen-addr", ":8086")
	setDefaultAndFlagBinding("server.listen_addr_tls", "listen-addr-tls", ":8443")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	resolveConfigPaths(
		"trust_dir",
		"remote_server.root_ca",
		"remote_server.tls_client_cert",
		"remote_server.tls_client_key",
	)
}

func setDefaultAndFlagBinding(key, flag string, value interface{}) {
	viper.SetDefault(key, value)
	viper.BindPFlag(key, rootCmd.PersistentFlags().Lookup(flag))
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(v VersionInfo) {
	version = v
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
