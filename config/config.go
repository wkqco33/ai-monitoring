package config

import (
	"os"
	"path/filepath"
	"time"

	"wconf"
)

// AppConfig 설정 정보
type AppConfig struct {
	CheckInterval   time.Duration `wconf:"check_interval" default:"10s"`
	CPUThreshold    float64       `wconf:"cpu_threshold" default:"90.0"`
	MemoryThreshold float64       `wconf:"memory_threshold" default:"90.0"`
	AzureEndpoint   string        `wconf:"azure_endpoint"`
	AzureOpenAIKey  string        `wconf:"azure_api_key"`
	AzureDeployment string        `wconf:"azure_deployment"`
	BotToken        string        `wconf:"bot_token"`
	CooldownPeriod  time.Duration `wconf:"cooldown_period" default:"5m"`
}

// GlobalConfig 전역 설정 인스턴스
var GlobalConfig = &AppConfig{}

// Load 설정을 로드합니다.
func Load(configPath string) error {
	options := []wconf.Option{
		wconf.WithEnv(),
		wconf.WithPrefix("AI_MONITORING"),
	}

	if configPath != "" {
		options = append(options, wconf.WithFiles(configPath))
	}

	return wconf.Load(GlobalConfig, options...)
}

// GetDefaultConfigPath 기본 설정 파일 경로를 반환합니다. (~/.config/ai-monitoring/config.yaml)
func GetDefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "config.yaml"
	}
	return filepath.Join(home, ".config", "ai-monitoring", "config.yaml")
}
