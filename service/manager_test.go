package service

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// newTestManager는 임시 경로와 목 의존성으로 Manager를 구성합니다.
func newTestManager(t *testing.T, run func(name string, args ...string) (string, error)) (*Manager, string) {
	t.Helper()
	dir := t.TempDir()
	m := &Manager{
		ServiceName:  "ai-monitoring",
		UnitPath:     filepath.Join(dir, "ai-monitoring.service"),
		EnvFile:      filepath.Join(dir, "ai-monitoring"),
		LogFile:      filepath.Join(dir, "ai-monitoring.log"),
		run:          run,
		geteuid:      func() int { return 0 },
		checkSystemd: func() error { return nil },
	}
	return m, dir
}

func testOptions() UnitOptions {
	return UnitOptions{
		Description: "AI PC Monitoring System",
		BinaryPath:  "/usr/local/bin/ai-monitoring",
		UserName:    "testuser",
		GroupName:   "testuser",
		Uid:         -1,
		Gid:         -1,
		EnvFile:     "/etc/default/ai-monitoring",
		LogFile:     "/var/log/ai-monitoring.log",
	}
}

func TestInstallWritesUnitFileAndEnablesService(t *testing.T) {
	var cmds [][]string
	m, dir := newTestManager(t, func(name string, args ...string) (string, error) {
		cmds = append(cmds, append([]string{name}, args...))
		return "", nil
	})

	if err := m.Install(testOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(m.UnitPath)
	if err != nil {
		t.Fatalf("unit file should be written: %v", err)
	}
	if !strings.Contains(string(data), "ExecStart=/usr/local/bin/ai-monitoring start") {
		t.Errorf("unit file should contain ExecStart, got:\n%s", data)
	}
	if !strings.Contains(string(data), "User=testuser") {
		t.Errorf("unit file should contain resolved user, got:\n%s", data)
	}

	if len(cmds) != 2 {
		t.Fatalf("expected 2 systemctl calls, got %v", cmds)
	}
	if cmds[0][1] != "daemon-reload" {
		t.Errorf("first command should be daemon-reload, got %v", cmds[0])
	}
	if cmds[1][1] != "enable" || cmds[1][2] != "--now" || cmds[1][3] != "ai-monitoring" {
		t.Errorf("second command should be enable --now, got %v", cmds[1])
	}

	if _, err := os.Stat(filepath.Join(dir, "ai-monitoring.log")); err != nil {
		t.Errorf("log file should be created: %v", err)
	}
}

func TestInstallCreatesEnvFileWhenMissing(t *testing.T) {
	m, _ := newTestManager(t, func(name string, args ...string) (string, error) { return "", nil })

	if err := m.Install(testOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(m.EnvFile)
	if err != nil {
		t.Fatalf("env file should be created: %v", err)
	}
	if !strings.Contains(string(data), "AI_MONITORING_LLM_PROVIDER") {
		t.Errorf("env file should contain template, got:\n%s", data)
	}
}

func TestInstallKeepsExistingEnvFile(t *testing.T) {
	m, _ := newTestManager(t, func(name string, args ...string) (string, error) { return "", nil })
	existing := "AI_MONITORING_API_KEY=secret-do-not-touch"
	if err := os.WriteFile(m.EnvFile, []byte(existing), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := m.Install(testOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(m.EnvFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != existing {
		t.Errorf("existing env file must not be overwritten, got:\n%s", data)
	}
}

func TestInstallBacksUpChangedUnitFile(t *testing.T) {
	m, _ := newTestManager(t, func(name string, args ...string) (string, error) { return "", nil })
	old := "[Unit]\nDescription=old\n"
	if err := os.WriteFile(m.UnitPath, []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := m.Install(testOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	backup, err := os.ReadFile(m.UnitPath + ".bak")
	if err != nil {
		t.Fatalf("backup should be created: %v", err)
	}
	if string(backup) != old {
		t.Errorf("backup should keep old content, got:\n%s", backup)
	}
}

func TestInstallNoBackupWhenIdentical(t *testing.T) {
	m, _ := newTestManager(t, func(name string, args ...string) (string, error) { return "", nil })
	opts := testOptions()
	if err := os.WriteFile(m.UnitPath, []byte(RenderUnit(opts)), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := m.Install(opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(m.UnitPath + ".bak"); !os.IsNotExist(err) {
		t.Error("backup should not be created for identical unit file")
	}
}

func TestInstallRequiresRoot(t *testing.T) {
	m, _ := newTestManager(t, func(name string, args ...string) (string, error) { return "", nil })
	m.geteuid = func() int { return 1000 }

	if err := m.Install(testOptions()); err == nil {
		t.Fatal("expected error when not running as root")
	}
}

func TestInstallRequiresSystemd(t *testing.T) {
	m, _ := newTestManager(t, func(name string, args ...string) (string, error) { return "", nil })
	m.checkSystemd = func() error { return errors.New("no systemctl") }

	if err := m.Install(testOptions()); err == nil {
		t.Fatal("expected error when systemd is unavailable")
	}
}

func TestUninstallSequence(t *testing.T) {
	var cmds [][]string
	m, _ := newTestManager(t, func(name string, args ...string) (string, error) {
		cmds = append(cmds, append([]string{name}, args...))
		return "", nil
	})
	if err := os.WriteFile(m.UnitPath, []byte(RenderUnit(testOptions())), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := m.Uninstall(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cmds) != 2 {
		t.Fatalf("expected 2 systemctl calls, got %v", cmds)
	}
	if cmds[0][1] != "disable" {
		t.Errorf("first command should be disable, got %v", cmds[0])
	}
	if cmds[1][1] != "daemon-reload" {
		t.Errorf("second command should be daemon-reload, got %v", cmds[1])
	}
	if _, err := os.Stat(m.UnitPath); !os.IsNotExist(err) {
		t.Error("unit file should be removed")
	}
}

func TestUninstallMissingServiceSucceeds(t *testing.T) {
	m, _ := newTestManager(t, func(name string, args ...string) (string, error) {
		// 서비스가 등록되지 않은 경우 disable만 실패합니다.
		if len(args) > 0 && args[0] == "disable" {
			return "", errors.New("service not found")
		}
		return "", nil
	})
	if err := m.Uninstall(); err != nil {
		t.Errorf("uninstall of missing service should succeed, got %v", err)
	}
}

func TestActionErrorContainsOutput(t *testing.T) {
	m, _ := newTestManager(t, func(name string, args ...string) (string, error) {
		return "Access denied", errors.New("exit status 1")
	})

	err := m.Start()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "Access denied") {
		t.Errorf("error should contain command output, got %v", err)
	}
	if !strings.Contains(err.Error(), "sudo") {
		t.Errorf("error should hint sudo usage, got %v", err)
	}
}

func TestStatusTreatsInactiveAsValid(t *testing.T) {
	m, _ := newTestManager(t, func(name string, args ...string) (string, error) {
		return "inactive", &exec.ExitError{}
	})

	out, err := m.Status()
	if err != nil {
		t.Fatalf("inactive service is a valid status: %v", err)
	}
	if out != "inactive" {
		t.Errorf("unexpected status output: %q", out)
	}
}

func TestStatusRealError(t *testing.T) {
	m, _ := newTestManager(t, func(name string, args ...string) (string, error) {
		return "", errors.New("systemd not running")
	})

	if _, err := m.Status(); err == nil {
		t.Fatal("expected error")
	}
}
