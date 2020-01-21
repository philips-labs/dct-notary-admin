package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/mitchellh/go-homedir"
	"github.com/philips-labs/dct-notary-admin/lib"
	"github.com/philips-labs/dct-notary-admin/lib/notary"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	configCmd = &cobra.Command{
		Use:   "config",
		Short: "get current configuration",
		Run: func(cmd *cobra.Command, _ []string) {
			cmd.Print(sprintConfig())
		},
	}
)

func init() {
	rootCmd.AddCommand(configCmd)
}

func sprintConfig() string {
	sb := new(strings.Builder)
	sb.WriteString("\nconfig:\n")
	writeSettings(sb, viper.AllSettings())
	return sb.String()
}

func writeSettings(w io.Writer, settings map[string]interface{}) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	writeRecurseSetting(tw, "", settings)
	tw.Flush()
}

func writeRecurseSetting(w io.Writer, prefix string, settings map[string]interface{}) {
	// maps can not guarantee same order, below is a trick to always print settings in same order.
	keys := make([]string, 0, len(settings))
	for k := range settings {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := settings[k]
		t := reflect.TypeOf(v)

		var key string
		if prefix != "" {
			key = prefix + "." + k
		} else {
			key = k
		}

		if t.Kind() == reflect.Map {
			writeRecurseSetting(w, key, v.(map[string]interface{}))
		} else {
			if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
				slice := strings.Replace(fmt.Sprintf("%q", v), "\" \"", "\", \"", -1)
				fmt.Fprintf(w, "\t%s:\t%s\n", key, slice)
			} else {
				fmt.Fprintf(w, "\t%s:\t%v\n", key, v)
			}
		}
	}
}

func unmarshalServerConfig() (*lib.ServerConfig, error) {
	var serverCfg lib.ServerConfig
	if err := viper.UnmarshalKey("server", &serverCfg); err != nil {
		return nil, err
	}
	return &serverCfg, nil
}

func unmarshalNotaryConfig() (*notary.Config, error) {
	var notaryCfg notary.Config
	if err := viper.Unmarshal(&notaryCfg); err != nil {
		return nil, err
	}
	return &notaryCfg, nil
}

func resolveConfigPaths(configKeys ...string) {
	for _, key := range configKeys {
		absolutePath, _ := homedir.Expand(viper.GetString(key))
		absolutePath = getPathRelativeToConfig(absolutePath)
		absolutePath = pathRelativeToCwd(absolutePath)
		viper.Set(key, absolutePath)
	}
}

func getPathRelativeToConfig(p string) string {
	configFile := viper.ConfigFileUsed()
	if p == "" || filepath.IsAbs(p) {
		return p
	}
	return filepath.Clean(filepath.Join(filepath.Dir(configFile), p))
}

func pathRelativeToCwd(path string) string {
	if path == "" || filepath.IsAbs(path) {
		return path
	}
	cwd, err := os.Getwd()
	if err != nil {
		return filepath.Clean(path)
	}
	return filepath.Clean(filepath.Join(cwd, path))
}
