package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/wkqco33/wcli/config"
)

// AppConfig 설정 정보
type AppConfig struct {
	CheckInterval   time.Duration `wconf:"check_interval" default:"10s"`
	CPUThreshold    float64       `wconf:"cpu_threshold" default:"90.0"`
	MemoryThreshold float64       `wconf:"memory_threshold" default:"90.0"`
	AzureEndpoint   string        `wconf:"azure_endpoint"`
	AzureOpenAIKey  string        `wconf:"azure_api_key"`
	AzureDeployment string        `wconf:"azure_deployment"`
	LLMProvider     string        `wconf:"llm_provider" default:"azure"`
	OllamaEndpoint  string        `wconf:"ollama_endpoint" default:"http://localhost:11434"`
	OllamaModel     string        `wconf:"ollama_model" default:"llama3"`
	BotToken        string        `wconf:"bot_token"`
	CooldownPeriod  time.Duration `wconf:"cooldown_period" default:"5m"`
}

// GlobalConfig 전역 설정 인스턴스
var GlobalConfig = &AppConfig{}

// Load 설정을 로드합니다.
func Load(configPath string) error {
	options := []config.BindOption{
		config.WithEnv(),
		config.WithPrefix("PCAM"),
		// 구조체 필드가 `wconf:"..."` 태그를 사용하므로 태그명을 유지합니다.
		config.WithTag("wconf"),
	}

	if configPath != "" {
		options = append(options, config.WithFiles(configPath))
	}

	return config.Load(GlobalConfig, options...)
}

// GetDefaultConfigPath 기본 설정 파일 경로를 반환합니다. (~/.config/pcam/config.yaml)
func GetDefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "config.yaml"
	}
	return filepath.Join(home, ".config", "pcam", "config.yaml")
}
