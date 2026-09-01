package cmd

import (
	"strings"

	"ai-monitoring/analyzer"
	"ai-monitoring/config"
	"ai-monitoring/logger"
	"github.com/wkqco33/wcli"
	"github.com/wkqco33/wcli/rich"
)

var analyzeLines int

var analyzeCmd = &wcli.Command{
	Use:   "analyze",
	Short: "Analyze recent logs for anomalies",
	Long:  "Send recent pcam logs to the LLM and summarize anomalies.",
	Run: func(ctx *wcli.Context) error {
		lines, err := logger.ReadRecent(analyzeLines)
		if err != nil {
			rich.Println("[yellow]No logs available at %s: %v[/yellow]", logger.LogFilePath(), err)
			return nil
		}
		if len(lines) == 0 {
			rich.Println("[yellow]No log entries found at %s[/yellow]", logger.LogFilePath())
			return nil
		}

		rich.Println("[cyan]Analyzing %d recent log entries with LLM...[/cyan]", len(lines))
		analysis, err := analyzer.AnalyzeRecentLogs(ctx, config.GlobalConfig, strings.Join(lines, "\n"))
		if err != nil {
			return err
		}

		rich.Println("[bold][white]Log Analysis:[/white][/bold]")
		rich.Println("[white]%s[/white]", analysis)
		return nil
	},
}

func init() {
	analyzeCmd.Flags().IntVar(&analyzeLines, "lines", "n", 100, "Number of recent log lines to analyze")
	rootCmd.AddCommand(analyzeCmd)
}
