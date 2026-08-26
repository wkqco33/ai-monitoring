package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// unsetEnv는 테스트가 호스트 환경변수의 영향을 받지 않도록 정리합니다.
func unsetEnv(t *testing.T) {
	t.Helper()
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "AI_MONITORING_") {
			key := strings.SplitN(e, "=", 2)[0]
			t.Setenv(key, "")
			os.Unsetenv(key)
		}
	}
}

func TestLoadFromYAML(t *testing.T) {
	unsetEnv(t)

	path := filepath.Join(t.TempDir(), "config.yaml")
	yaml := `
check_interval: 5s
cpu_threshold: 80.0
memory_threshold: 75.5
llm_provider: ollama
`
	if err := os.WriteFile(path, []byte(yaml), 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	if err := Load(path); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if GlobalConfig.CheckInterval != 5*time.Second {
		t.Errorf("CheckInterval = %v, want 5s", GlobalConfig.CheckInterval)
	}
	if GlobalConfig.CPUThreshold != 80.0 {
		t.Errorf("CPUThreshold = %v, want 80.0", GlobalConfig.CPUThreshold)
	}
	if GlobalConfig.MemoryThreshold != 75.5 {
		t.Errorf("MemoryThreshold = %v, want 75.5", GlobalConfig.MemoryThreshold)
	}
	if GlobalConfig.LLMProvider != "ollama" {
		t.Errorf("LLMProvider = %v, want ollama", GlobalConfig.LLMProvider)
	}
}

func TestLoadDefaultsWhenNoSource(t *testing.T) {
	unsetEnv(t)

	// 설정 파일도 환경변수도 없으면 기본값이 사용되어야 합니다.
	if err := Load(""); err != nil {
		t.Fatalf("Load() with no source error = %v", err)
	}

	if GlobalConfig.CheckInterval != 10*time.Second {
		t.Errorf("CheckInterval = %v, want default 10s", GlobalConfig.CheckInterval)
	}
	if GlobalConfig.CPUThreshold != 90.0 {
		t.Errorf("CPUThreshold = %v, want default 90.0", GlobalConfig.CPUThreshold)
	}
	if GlobalConfig.LLMProvider != "azure" {
		t.Errorf("LLMProvider = %v, want default azure", GlobalConfig.LLMProvider)
	}
}

func TestLoadFromEnvOverridesFile(t *testing.T) {
	unsetEnv(t)

	t.Setenv("AI_MONITORING_LLM_PROVIDER", "ollama")
	t.Setenv("AI_MONITORING_CPU_THRESHOLD", "70.0")

	path := filepath.Join(t.TempDir(), "config.yaml")
	yaml := "llm_provider: azure\n"
	if err := os.WriteFile(path, []byte(yaml), 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	if err := Load(path); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if GlobalConfig.LLMProvider != "ollama" {
		t.Errorf("LLMProvider = %v, want env ollama to override file", GlobalConfig.LLMProvider)
	}
	if GlobalConfig.CPUThreshold != 70.0 {
		t.Errorf("CPUThreshold = %v, want env 70.0", GlobalConfig.CPUThreshold)
	}
}
