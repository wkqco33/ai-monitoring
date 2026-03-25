package analyzer

import (
	"strings"
	"testing"

	"ai-monitoring/monitor"
)

func TestFormatPrompt(t *testing.T) {
	state := &monitor.SystemState{
		CPUUsage: 95.5,
		MemUsage: 80.2,
		Processes: []monitor.ProcessInfo{
			{PID: 1234, Name: "chrome", CPUProg: 40.0, MemProg: 15.0},
			{PID: 5678, Name: "vscode", CPUProg: 30.0, MemProg: 10.0},
		},
	}

	prompt := formatPrompt(state)

	if !strings.Contains(prompt, "95.50%") {
		t.Errorf("Expected prompt to contain CPU usage, got: %s", prompt)
	}
	if !strings.Contains(prompt, "chrome") {
		t.Errorf("Expected prompt to contain chrome, got: %s", prompt)
	}
}
