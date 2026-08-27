package logger

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadRecent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.log")
	if err := os.WriteFile(path, []byte("l1\nl2\nl3\nl4\nl5\n"), 0o644); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}

	got, err := readRecent(path, 3)
	if err != nil {
		t.Fatalf("readRecent() error = %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("len(got) = %d, want 3", len(got))
	}
	if got[0] != "l3" || got[2] != "l5" {
		t.Errorf("got = %v, want last 3 lines", got)
	}
}

func TestReadRecentMoreThanLines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.log")
	if err := os.WriteFile(path, []byte("a\nb\n"), 0o644); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}

	got, err := readRecent(path, 10)
	if err != nil {
		t.Fatalf("readRecent() error = %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len(got) = %d, want 2", len(got))
	}
}

func TestLogFileCreatedOnInit(t *testing.T) {
	old := logFilePath
	t.Cleanup(func() {
		logFilePath = old
		if logFile != nil {
			logFile.Close()
		}
	})
	logFilePath = filepath.Join(t.TempDir(), "test.log")

	Init(false)

	if _, err := os.Stat(logFilePath); err != nil {
		t.Fatalf("log file not created: %v", err)
	}

	slog.Info("hello")
	lines, err := ReadRecent(10)
	if err != nil {
		t.Fatalf("ReadRecent() error = %v", err)
	}
	found := false
	for _, l := range lines {
		if strings.Contains(l, "hello") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("ReadRecent() did not contain logged message, got %q", lines)
	}
}

func TestReadRecentMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nope.log")
	if _, err := readRecent(path, 5); err == nil {
		t.Fatal("readRecent() on missing file: expected error")
	}
}
