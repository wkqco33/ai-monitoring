package logger

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

// logFilePath 로그 파일 경로. 테스트에서 우회할 수 있도록 패키지 변수로 유지합니다.
var logFilePath string

// logFile 열려 있는 로그 파일 핸들 (애플리케이션 수명 동안 유지)
var logFile io.Closer

// Init 전역 로거를 초기화하고 표준 출력과 로그 파일에 동시에 기록합니다.
func Init(debug bool) {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{Level: level}
	stdoutHandler := slog.NewTextHandler(os.Stdout, opts)

	file, err := openLogFile()
	if err == nil {
		logFile = file
		fileHandler := slog.NewTextHandler(file, opts)
		slog.SetDefault(slog.New(slog.NewMultiHandler(stdoutHandler, fileHandler)))
		return
	}
	slog.SetDefault(slog.New(stdoutHandler))
}

// LogFilePath 기본 로그 파일 경로를 반환합니다. (~/.local/state/ai-monitoring/ai-monitoring.log)
func LogFilePath() string {
	if logFilePath != "" {
		return logFilePath
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "ai-monitoring.log"
	}
	return filepath.Join(home, ".local", "state", "ai-monitoring", "ai-monitoring.log")
}

// openLogFile 로그 파일을 추가 쓰기 모드로 엽니다.
func openLogFile() (*os.File, error) {
	path := LogFilePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	return os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
}

// ReadRecent 기본 로그 파일에서 최근 n줄을 읽어 반환합니다.
func ReadRecent(n int) ([]string, error) {
	return readRecent(LogFilePath(), n)
}

// readRecent 지정 경로의 로그 파일에서 마지막 n줄을 읽어 반환합니다.
func readRecent(path string, n int) ([]string, error) {
	if n <= 0 {
		return nil, fmt.Errorf("n must be positive, got %d", n)
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > n {
			lines = lines[1:]
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}
