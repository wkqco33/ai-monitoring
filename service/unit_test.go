package service

import (
	"strings"
	"testing"
)

func TestRenderUnit(t *testing.T) {
	opts := UnitOptions{
		Description: "AI PC Monitoring System",
		BinaryPath:  "/usr/local/bin/pcam",
		UserName:    "seoyc",
		GroupName:   "seoyc",
		EnvFile:     "/etc/default/pcam",
		LogFile:     "/var/log/pcam.log",
	}

	got := RenderUnit(opts)

	want := []string{
		"Description=AI PC Monitoring System",
		"User=seoyc",
		"Group=seoyc",
		"EnvironmentFile=-/etc/default/pcam",
		"ExecStart=/usr/local/bin/pcam start",
		"StandardOutput=append:/var/log/pcam.log",
		"StandardError=append:/var/log/pcam.log",
		"WantedBy=multi-user.target",
		"Restart=always",
	}
	for _, w := range want {
		if !strings.Contains(got, w) {
			t.Errorf("unit file should contain %q\nunit:\n%s", w, got)
		}
	}
}

func TestRenderUnitDeterministic(t *testing.T) {
	opts := UnitOptions{
		Description: "d",
		BinaryPath:  "/bin/x",
		UserName:    "u",
		GroupName:   "g",
		EnvFile:     "/env",
		LogFile:     "/log",
	}
	if RenderUnit(opts) != RenderUnit(opts) {
		t.Error("RenderUnit should be deterministic")
	}
}
