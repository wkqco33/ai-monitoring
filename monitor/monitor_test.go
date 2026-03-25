package monitor

import (
	"testing"
)

func TestGetSystemState(t *testing.T) {
	state, err := GetSystemState(5)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if state == nil {
		t.Fatalf("Expected state to not be nil")
	}

	if state.CPUUsage < 0 || state.CPUUsage > 100 {
		t.Errorf("Invalid CPU usage: %f", state.CPUUsage)
	}

	if state.MemUsage < 0 || state.MemUsage > 100 {
		t.Errorf("Invalid Mem usage: %f", state.MemUsage)
	}

	if len(state.Processes) > 5 {
		t.Errorf("Expected at most 5 processes, got %d", len(state.Processes))
	}
}
