package cmd

import (
	"fmt"
	"os"
	"strings"

	"ai-monitoring/service"
	"github.com/wkqco33/wcli"
	"github.com/wkqco33/wcli/rich"
)

var serviceCmd = &wcli.Command{
	Use:   "service",
	Short: "Manage the ai-monitoring systemd service",
	Long:  "Install and control the ai-monitoring systemd service (Linux only). Most operations require root.",
}

var serviceInstallCmd = &wcli.Command{
	Use:   "install",
	Short: "Install and enable the systemd service",
	Run: func(ctx *wcli.Context) error {
		opts, err := service.ResolveInstallOptions()
		if err != nil {
			return err
		}
		if opts.BinaryPath != service.DefaultBinaryPath {
			rich.Println("[yellow]Warning: installing from %s. Prefer 'task install' to place the binary at %s first.[/yellow]",
				opts.BinaryPath, service.DefaultBinaryPath)
		}
		if err := service.New().Install(opts); err != nil {
			return err
		}
		rich.Println("[green]Service installed, started and enabled on boot.[/green]")
		rich.Println("[dim]Env file: %s — fill in API keys, then run 'ai-monitoring service restart'.[/dim]", service.DefaultEnvFile)
		return nil
	},
}

var serviceUninstallCmd = &wcli.Command{
	Use:   "uninstall",
	Short: "Stop, disable and remove the systemd service",
	Run: func(ctx *wcli.Context) error {
		if err := service.New().Uninstall(); err != nil {
			return err
		}
		rich.Println("[green]Service uninstalled. User config and env file are kept.[/green]")
		return nil
	},
}

var serviceStatusCmd = &wcli.Command{
	Use:   "status",
	Short: "Show the systemd service status",
	Run: func(ctx *wcli.Context) error {
		out, err := service.New().Status()
		if err != nil {
			return err
		}
		if strings.TrimSpace(out) == "" {
			fmt.Fprintln(os.Stdout, "No status output from systemctl.")
			return nil
		}
		fmt.Fprint(os.Stdout, out)
		return nil
	},
}

var serviceStartCmd = &wcli.Command{
	Use:   "start",
	Short: "Start the systemd service",
	Run: func(ctx *wcli.Context) error {
		if err := service.New().Start(); err != nil {
			return err
		}
		rich.Println("[green]Service started.[/green]")
		return nil
	},
}

var serviceStopCmd = &wcli.Command{
	Use:   "stop",
	Short: "Stop the systemd service",
	Run: func(ctx *wcli.Context) error {
		if err := service.New().Stop(); err != nil {
			return err
		}
		rich.Println("[green]Service stopped.[/green]")
		return nil
	},
}

var serviceRestartCmd = &wcli.Command{
	Use:   "restart",
	Short: "Restart the systemd service",
	Run: func(ctx *wcli.Context) error {
		if err := service.New().Restart(); err != nil {
			return err
		}
		rich.Println("[green]Service restarted.[/green]")
		return nil
	},
}

func init() {
	serviceCmd.AddCommand(
		serviceInstallCmd,
		serviceUninstallCmd,
		serviceStatusCmd,
		serviceStartCmd,
		serviceStopCmd,
		serviceRestartCmd,
	)
	rootCmd.AddCommand(serviceCmd)
}
