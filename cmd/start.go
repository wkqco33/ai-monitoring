package cmd

import (
	"context"
	"log/slog"
	"time"

	"github.com/seoyc/wcli"
	"github.com/seoyc/wcli/rich"
	"ai-monitoring/analyzer"
	"ai-monitoring/config"
	"ai-monitoring/monitor"
	"ai-monitoring/notifier"
)

var (
	interval        time.Duration
	cpuThreshold    float64
	memThreshold    float64
	cooldownPeriod  time.Duration
	azureEndpoint   string
	azureOpenAIKey  string
	azureDeployment string
)

var startCmd = &wcli.Command{
	Use:   "start",
	Short: "Start monitoring",
	Run: func(ctx *wcli.Context) error {
		cfg := config.GlobalConfig

		// 플래그가 명시적으로 설정된 경우 설정을 덮어씁니다.
		if ctx.IsSet("interval") {
			cfg.CheckInterval = interval
		}
		if ctx.IsSet("cpu") {
			cfg.CPUThreshold = cpuThreshold
		}
		if ctx.IsSet("mem") {
			cfg.MemoryThreshold = memThreshold
		}
		if ctx.IsSet("cooldown") {
			cfg.CooldownPeriod = cooldownPeriod
		}
		if ctx.IsSet("azure-endpoint") {
			cfg.AzureEndpoint = azureEndpoint
		}
		if ctx.IsSet("azure-key") {
			cfg.AzureOpenAIKey = azureOpenAIKey
		}
		if ctx.IsSet("azure-deployment") {
			cfg.AzureDeployment = azureDeployment
		}

		rich.Println("[bold][green]Starting ai-monitoring[/green][/bold]")
		rich.Println("[dim]Interval: %v, CPU Threshold: %.1f%%, Memory Threshold: %.1f%%[/dim]", 
			cfg.CheckInterval, cfg.CPUThreshold, cfg.MemoryThreshold)

		// 부팅 로그 진단
		rich.Println("[cyan]Checking system boot logs...[/cyan]")
		bootLogs, err := monitor.GetBootLogs()
		if err != nil {
			rich.Println("[yellow]Warning: Could not fetch boot logs: %v[/yellow]", err)
		} else {
			summary := monitor.GetBootSummary(bootLogs)
			rich.Println("[dim]Recent Boot Issues:\n%s[/dim]", summary)
			
			rich.Println("[magenta]Analyzing boot logs with LLM...[/magenta]")
			bootAnalysis, err := analyzer.AnalyzeBootLogs(context.Background(), cfg, bootLogs)
			if err != nil {
				rich.Println("[red]Boot diagnosis failed: %v[/red]", err)
			} else {
				rich.Println("[bold][white]Boot Diagnosis:[/white][/bold]")
				rich.Println("[white]%s[/white]", bootAnalysis)
			}
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
			lastAlert = time.Now()

			analysis, err := analyzer.AnalyzeSystemState(runCtx, cfg, state)
			if err != nil {
				rich.Println("[red]Failed to analyze system state: %v[/red]", err)
				continue
			}

			rich.Println("[bold][cyan]Analysis complete[/cyan][/bold]")
			rich.Println("[white]%s[/white]", analysis)
			
			notifier.Notify(cfg, "PC 상태 이상 감지", analysis)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	
	f := startCmd.Flags()
	f.DurationVar(&interval, "interval", "i", 10*time.Second, "Check interval")
	f.Float64Var(&cpuThreshold, "cpu", "c", 90.0, "CPU usage threshold (%)")
	f.Float64Var(&memThreshold, "mem", "m", 90.0, "Memory usage threshold (%)")
	f.DurationVar(&cooldownPeriod, "cooldown", "C", 5*time.Minute, "Cooldown period for alerts")
	f.StringVar(&azureEndpoint, "azure-endpoint", "e", "", "Azure OpenAI Endpoint URL")
	f.StringVar(&azureOpenAIKey, "azure-key", "k", "", "Azure OpenAI API Key")
	f.StringVar(&azureDeployment, "azure-deployment", "D", "", "Azure OpenAI Deployment Name")
}
