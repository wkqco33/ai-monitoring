package cmd

import (
	"fmt"
	"os"

	"github.com/wkqco33/wcli"
	"github.com/wkqco33/wcli/rich"

	"ai-monitoring/config"
)

// resolveConfigPath 설정 경로를 결정합니다.
func resolveConfigPath() string {
	if configPath != "" {
		return configPath
	}
	return config.GetDefaultConfigPath()
}

var configCmd = &wcli.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  "Initialize, display, or update the configuration file.",
}

var configInitCmd = &wcli.Command{
	Use:   "init",
	Short: "Create a default configuration file",
	Run: func(ctx *wcli.Context) error {
		path := resolveConfigPath()
		if err := config.WriteConfig(path); err != nil {
			return err
		}
		rich.Println("[green]Configuration file created: %s[/green]", path)
		return nil
	},
}

var configShowCmd = &wcli.Command{
	Use:   "show",
	Short: "Display the current configuration",
	Run: func(ctx *wcli.Context) error {
		data, err := config.DumpConfig()
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, data)
		return nil
	},
}

var configSetCmd = &wcli.Command{
	Use:   "set <key> <value>",
	Short: "Update a configuration value and save it",
	Run: func(ctx *wcli.Context) error {
		if len(ctx.Args) != 2 {
			return fmt.Errorf("usage: ai-monitoring config set <key> <value>")
		}
		key, value := ctx.Args[0], ctx.Args[1]
		if err := config.SetConfig(key, value); err != nil {
			return err
		}
		path := resolveConfigPath()
		if err := config.WriteConfig(path); err != nil {
			return err
		}
		rich.Println("[config] [green]%s[/green] set to [bold]%s[/bold] and saved to %s", key, value, path)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configInitCmd, configShowCmd, configSetCmd)
	rootCmd.AddCommand(configCmd)
}
