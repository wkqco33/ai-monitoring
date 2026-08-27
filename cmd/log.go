package cmd

import (
	"fmt"
	"os"

	"github.com/wkqco33/wcli"

	"ai-monitoring/logger"
)

var (
	logLines int
)

var logCmd = &wcli.Command{
	Use:   "log",
	Short: "Show recent log entries",
	Run: func(ctx *wcli.Context) error {
		if len(ctx.Args) > 0 {
			return fmt.Errorf("unexpected argument: %s", ctx.Args[0])
		}
		lines, err := logger.ReadRecent(logLines)
		if err != nil {
			return err
		}
		if len(lines) == 0 {
			fmt.Fprintf(os.Stdout, "No log entries found at %s\n", logger.LogFilePath())
			return nil
		}
		fmt.Fprintf(os.Stdout, "Recent logs (%s)\n", logger.LogFilePath())
		for _, line := range lines {
			fmt.Fprintln(os.Stdout, line)
		}
		return nil
	},
}

func init() {
	logCmd.Flags().IntVar(&logLines, "lines", "n", 50, "Number of recent log lines to show")
	rootCmd.AddCommand(logCmd)
}
