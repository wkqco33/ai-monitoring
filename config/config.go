package config

import (
	"time"
)

// AppConfig 설정 정보
type AppConfig struct {
	CheckInterval   time.Duration
	CPUThreshold    float64
	MemoryThreshold float64
	AzureEndpoint   string
	AzureOpenAIKey  string
	AzureDeployment string
	BotToken        string
	CooldownPeriod  time.Duration
}

// DefaultConfig 기본 설정 반환
func DefaultConfig() *AppConfig {
	return &AppConfig{
		CheckInterval:   10 * time.Second,
		CPUThreshold:    90.0,
		MemoryThreshold: 90.0,
		CooldownPeriod:  5 * time.Minute,
	}
}
