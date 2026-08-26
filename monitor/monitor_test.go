package monitor

import (
	"strings"
	"testing"
)

func TestGetBootSummary(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", "탐지된 부팅 에러나 특이사항이 없습니다."},
		{"whitespace", "   \n  ", "탐지된 부팅 에러나 특이사항이 없습니다."},
		{"short", "line1", "line1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := GetBootSummary(tc.in); got != tc.want {
				t.Errorf("GetBootSummary(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestGetBootSummaryTruncatesLongLogs(t *testing.T) {
	lines := make([]string, 80)
	for i := range lines {
		lines[i] = "line"
	}
	got := GetBootSummary(strings.Join(lines, "\n"))

	if !strings.Contains(got, "...(중략)...") {
		t.Errorf("expected truncation marker, got: %s", got)
	}
	if got == strings.Join(lines, "\n") {
		t.Errorf("expected summary to be truncated")
	}
}

func TestGetUsage(t *testing.T) {
	cpuUsage, memUsage, err := GetUsage()
	if err != nil {
		t.Fatalf("GetUsage() error = %v", err)
	}
	if cpuUsage < 0 || cpuUsage > 100 {
		t.Errorf("invalid CPU usage: %f", cpuUsage)
	}
	if memUsage < 0 || memUsage > 100 {
		t.Errorf("invalid memory usage: %f", memUsage)
	}
}

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
