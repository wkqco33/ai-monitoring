package cmd

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"
	"ai-monitoring/analyzer"
	"ai-monitoring/config"
	"ai-monitoring/monitor"
	"ai-monitoring/notifier"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start monitoring",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.DefaultConfig()

		interval, _ := cmd.Flags().GetDuration("interval")
		if interval > 0 {
			cfg.CheckInterval = interval
		}

		cpu, _ := cmd.Flags().GetFloat64("cpu")
		if cpu > 0 {
			cfg.CPUThreshold = cpu
		}

		mem, _ := cmd.Flags().GetFloat64("mem")
		if mem > 0 {
			cfg.MemoryThreshold = mem
		}

		cooldown, _ := cmd.Flags().GetDuration("cooldown")
		if cooldown > 0 {
			cfg.CooldownPeriod = cooldown
		}

		cfg.AzureEndpoint, _ = cmd.Flags().GetString("azure-endpoint")
		cfg.AzureOpenAIKey, _ = cmd.Flags().GetString("azure-key")
		cfg.AzureDeployment, _ = cmd.Flags().GetString("azure-deployment")
		
		if cfg.AzureEndpoint == "" {
			cfg.AzureEndpoint = os.Getenv("AZURE_ENDPOINT")
		}
		if cfg.AzureOpenAIKey == "" {
			cfg.AzureOpenAIKey = os.Getenv("AZURE_API_KEY")
		}
		if cfg.AzureDeployment == "" {
			cfg.AzureDeployment = os.Getenv("AZURE_DEPLOYMENT")
		}

		slog.Info("Starting ai-monitoring", "interval", cfg.CheckInterval, "cpu_threshold", cfg.CPUThreshold, "mem_threshold", cfg.MemoryThreshold)

		triggerCh := make(chan *monitor.SystemState)
		go monitor.Start(cfg, triggerCh)

		var lastAlert time.Time

		ctx := context.Background()

		for state := range triggerCh {
			if time.Since(lastAlert) < cfg.CooldownPeriod {
				slog.Debug("Cooldown active, skipping analysis", "remaining", cfg.CooldownPeriod-time.Since(lastAlert))
				continue
			}

			slog.Info("Anomaly detected. Starting analysis...")
			lastAlert = time.Now()

			analysis, err := analyzer.AnalyzeSystemState(ctx, cfg, state)
			if err != nil {
				slog.Error("Failed to analyze system state", "error", err)
				continue
			}

			slog.Info("Analysis complete", "result", analysis)
			notifier.Notify(cfg, "PC 상태 이상 감지", analysis)
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().Duration("interval", 10*time.Second, "Check interval")
	startCmd.Flags().Float64("cpu", 90.0, "CPU usage threshold (%)")
	startCmd.Flags().Float64("mem", 90.0, "Memory usage threshold (%)")
	startCmd.Flags().Duration("cooldown", 5*time.Minute, "Cooldown period for alerts")
	startCmd.Flags().String("azure-endpoint", "", "Azure OpenAI Endpoint URL")
	startCmd.Flags().String("azure-key", "", "Azure OpenAI API Key")
	startCmd.Flags().String("azure-deployment", "", "Azure OpenAI Deployment Name")
}
