package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWriteConfigCreatesFile(t *testing.T) {
	unsetEnv(t)

	// GlobalConfig에 테스트 값 반영 후 임시 파일로 저장
	prev := *GlobalConfig
	t.Cleanup(func() { *GlobalConfig = prev })

	GlobalConfig.CheckInterval = 5 * time.Second
	GlobalConfig.CPUThreshold = 80.0
	GlobalConfig.LLMProvider = "ollama"

	path := filepath.Join(t.TempDir(), "sub", "config.yaml")
	if err := WriteConfig(path); err != nil {
		t.Fatalf("WriteConfig() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read written config: %v", err)
	}
	content := string(data)

	for _, want := range []string{
		"check_interval: 5s",
		"cpu_threshold: 80",
		"llm_provider: ollama",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("written config missing %q, got:\n%s", want, content)
		}
	}
}

func TestSetConfigString(t *testing.T) {
	unsetEnv(t)

	if err := SetConfig("llm_provider", "ollama"); err != nil {
		t.Fatalf("SetConfig() error = %v", err)
	}
	if GlobalConfig.LLMProvider != "ollama" {
		t.Errorf("LLMProvider = %q, want ollama", GlobalConfig.LLMProvider)
	}
}

func TestSetConfigFloat(t *testing.T) {
	unsetEnv(t)

	if err := SetConfig("cpu_threshold", "75.5"); err != nil {
		t.Fatalf("SetConfig() error = %v", err)
	}
	if GlobalConfig.CPUThreshold != 75.5 {
		t.Errorf("CPUThreshold = %v, want 75.5", GlobalConfig.CPUThreshold)
	}
}

func TestSetConfigDuration(t *testing.T) {
	unsetEnv(t)

	if err := SetConfig("cooldown_period", "3m"); err != nil {
		t.Fatalf("SetConfig() error = %v", err)
	}
	if GlobalConfig.CooldownPeriod != 3*time.Minute {
		t.Errorf("CooldownPeriod = %v, want 3m", GlobalConfig.CooldownPeriod)
	}
}

func TestSetConfigInvalidValue(t *testing.T) {
	unsetEnv(t)

	if err := SetConfig("cpu_threshold", "not-a-number"); err == nil {
		t.Fatal("SetConfig() with invalid float value: expected error")
	}
}

func TestSetConfigUnknownKey(t *testing.T) {
	unsetEnv(t)

	if err := SetConfig("no_such_key", "x"); err == nil {
		t.Fatal("SetConfig() with unknown key: expected error")
	}
}
