package monitor

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// GetBootLogs 운영체제별 부팅 로그를 가져옵니다.
func GetBootLogs() (string, error) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		// systemd 기반 리눅스에서 이번 부팅 로그의 마지막 100줄을 가져옵니다.
		cmd = exec.Command("journalctl", "-b", "-p", "err..emerg", "--no-pager", "-n", "100")
	case "darwin":
		// macOS에서 부팅 관련 로그를 가져옵니다.
		cmd = exec.Command("log", "show", "--predicate", "eventMessage contains \"boot\"", "--last", "1h")
	case "windows":
		// Windows 이벤트 로그에서 최근 시스템 에러 로그를 가져옵니다.
		powershellCmd := `Get-WinEvent -LogName System | Where-Object {$_.LevelDisplayName -eq "Error"} | Select-Object -First 20 | Format-Table -HideTableHeaders Message`
		cmd = exec.Command("powershell", "-Command", powershellCmd)
	default:
		return "", fmt.Errorf("지원하지 않는 운영체제입니다: %s", runtime.GOOS)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		// 명령어가 실패하더라도(예: journalctl 권한 문제 등) 에러 메시지를 포함해 반환합니다.
		return "", fmt.Errorf("로그를 가져오는데 실패했습니다: %v (출력: %s)", err, string(out))
	}

	return string(out), nil
}

// GetBootSummary 부팅 로그가 너무 길 경우를 대비해 요약된 텍스트를 생성합니다.
func GetBootSummary(logs string) string {
	if logs == "" || strings.TrimSpace(logs) == "" {
		return "탐지된 부팅 에러나 특이사항이 없습니다."
	}
	lines := strings.Split(logs, "\n")
	if len(lines) > 50 {
		return strings.Join(lines[:50], "\n") + "\n...(중략)..."
	}
	return logs
}
