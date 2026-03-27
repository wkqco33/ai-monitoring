package cmd

import (
	"os"

	"github.com/seoyc/wcli"
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

		// 설정 로드
		if err := config.Load(configPath); err != nil {
			// 설정 파일이 없어도 기본값으로 동작하도록 에러 처리는 로그만 남기거나 무시할 수 있습니다.
			// 여기서는 wconf.Load가 설정 소스가 없으면 에러를 내므로, 
			// configPath가 기본값일 때 파일이 없는 경우는 허용하도록 처리할 수 있습니다.
		}
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
