package cmd

import (
	"os"

	"github.com/wkqco33/wcli"

	"ai-monitoring/config"
	"ai-monitoring/logger"
)

var (
	debug      bool
	configPath string
)

var rootCmd = &wcli.Command{
	Use:   "ai-monitoring",
	Short: "AI PC Monitoring System",
	Long:  "Automatically monitors PC state and uses LLM to analyze anomalies.",
	PersistentPreRun: func(ctx *wcli.Context) error {
		logger.Init(debug)
		// 설정 파일이 없으면 기본값으로 동작하도록 로드 에러는 무시합니다.
		_ = config.Load(configPath)
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", "d", false, "Enable debug mode")
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "c", config.GetDefaultConfigPath(), "Path to configuration file")
}
