package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"ai-monitoring/logger"
)

var (
	debug bool
)

var rootCmd = &cobra.Command{
	Use:   "ai-monitoring",
	Short: "AI PC Monitoring System",
	Long:  "Automatically monitors PC state and uses LLM to analyze anomalies.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logger.Init(debug)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug mode")
}
