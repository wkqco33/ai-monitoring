package cmd

import (
	"context"
	"log/slog"
	"time"

	"ai-monitoring/analyzer"
	"ai-monitoring/config"
	"ai-monitoring/monitor"
	"ai-monitoring/notifier"
	"github.com/wkqco33/wcli"
	"github.com/wkqco33/wcli/rich"
)

var (
	interval        time.Duration
	cpuThreshold    float64
	memThreshold    float64
	cooldownPeriod  time.Duration
	azureEndpoint   string
	azureOpenAIKey  string
	azureDeployment string
	startFlags      *wcli.FlagSet
)

var startCmd = &wcli.Command{
	Use:   "start",
	Short: "Start monitoring",
	Run: func(ctx *wcli.Context) error {
		cfg := config.GlobalConfig

		// 플래그가 명시적으로 설정된 경우 설정을 덮어씁니다.
		if startFlags.Changed("interval") {
			cfg.CheckInterval = interval
		}
		if startFlags.Changed("cpu") {
			cfg.CPUThreshold = cpuThreshold
		}
		if startFlags.Changed("mem") {
			cfg.MemoryThreshold = memThreshold
		}
		if startFlags.Changed("cooldown") {
			cfg.CooldownPeriod = cooldownPeriod
		}
		if startFlags.Changed("azure-endpoint") {
			cfg.AzureEndpoint = azureEndpoint
		}
		if startFlags.Changed("azure-key") {
			cfg.AzureOpenAIKey = azureOpenAIKey
		}
		if startFlags.Changed("azure-deployment") {
			cfg.AzureDeployment = azureDeployment
		}

		rich.Println("[bold][green]Starting ai-monitoring[/green][/bold]")
		slog.Info("monitoring started",
			"interval", cfg.CheckInterval,
			"cpu_threshold", cfg.CPUThreshold,
			"memory_threshold", cfg.MemoryThreshold)
		rich.Println("[dim]Interval: %v, CPU Threshold: %.1f%%, Memory Threshold: %.1f%%[/dim]",
			cfg.CheckInterval, cfg.CPUThreshold, cfg.MemoryThreshold)

		// 부팅 로그 진단
		rich.Println("[cyan]Waiting for network to stabilize (5s)...[/cyan]")
		time.Sleep(5 * time.Second)

		analysis, err := runBootDiagnosis(ctx, cfg)
		if err != nil {
			rich.Println("[yellow]Warning: Boot diagnosis skipped: %v[/yellow]", err)
		} else if analysis == "" {
			rich.Println("[green]No boot issues detected.[/green]")
		} else {
			rich.Println("[bold][white]Boot Diagnosis:[/white][/bold]")
			rich.Println("[white]%s[/white]", analysis)
		}

		triggerCh := make(chan *monitor.SystemState)
		go monitor.Start(cfg, triggerCh)

		var lastAlert time.Time
		runCtx := context.Background()

		for state := range triggerCh {
			if time.Since(lastAlert) < cfg.CooldownPeriod {
				slog.Debug("Cooldown active, skipping analysis", "remaining", cfg.CooldownPeriod-time.Since(lastAlert))
				continue
			}

			rich.Println("[yellow]Anomaly detected. Starting analysis...[/yellow]")
			slog.Info("anomaly detected, starting analysis",
				"cpu", state.CPUUsage,
				"memory", state.MemUsage)
			lastAlert = time.Now()

			analysis, err := analyzer.AnalyzeSystemState(runCtx, cfg, state)
			if err != nil {
				rich.Println("[red]Failed to analyze system state: %v[/red]", err)
				slog.Error("failed to analyze system state", "error", err)
				continue
			}

			rich.Println("[bold][cyan]Analysis complete[/cyan][/bold]")
			slog.Info("analysis complete")
			rich.Println("[white]%s[/white]", analysis)

			notifier.Notify(cfg, "PC 상태 이상 감지", analysis)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	startFlags = startCmd.Flags()
	f := startFlags
	f.DurationVar(&interval, "interval", "i", 10*time.Second, "Check interval")
	f.Float64Var(&cpuThreshold, "cpu", "c", 90.0, "CPU usage threshold (%)")
	f.Float64Var(&memThreshold, "mem", "m", 90.0, "Memory usage threshold (%)")
	f.DurationVar(&cooldownPeriod, "cooldown", "C", 5*time.Minute, "Cooldown period for alerts")
	f.StringVar(&azureEndpoint, "azure-endpoint", "e", "", "Azure OpenAI Endpoint URL")
	f.StringVar(&azureOpenAIKey, "azure-key", "k", "", "Azure OpenAI API Key")
	f.StringVar(&azureDeployment, "azure-deployment", "D", "", "Azure OpenAI Deployment Name")
}
