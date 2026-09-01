package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ai-monitoring/analyzer"
	"ai-monitoring/config"
	"ai-monitoring/monitor"
	"github.com/wkqco33/wcli/rich"
)

const (
	bootMaxRetries    = 5
	bootRetryInterval = 5 * time.Second
)

// 부팅 진단의 외부 의존성. 테스트에서 교체할 수 있도록 패키지 변수로 유지합니다.
var (
	getBootLogsFn     = monitor.GetBootLogs
	analyzeBootLogsFn = analyzer.AnalyzeBootLogs
	bootRetryDelay    = time.Sleep
)

// runBootDiagnosis 부팅 로그를 수집하고 LLM으로 진단합니다.
// 로그가 비어 있으면 LLM 호출 없이 빈 문자열을 반환합니다.
func runBootDiagnosis(ctx context.Context, cfg *config.AppConfig) (string, error) {
	rich.Println("[cyan]Checking system boot logs...[/cyan]")
	logs, err := getBootLogsFn()
	if err != nil {
		return "", fmt.Errorf("부팅 로그 수집 실패: %w", err)
	}
	if strings.TrimSpace(logs) == "" {
		return "", nil
	}

	rich.Println("[dim]Recent Boot Issues:\n%s[/dim]", monitor.GetBootSummary(logs))
	rich.Println("[magenta]Analyzing boot logs with LLM...[/magenta]")

	var lastErr error
	for i := 0; i < bootMaxRetries; i++ {
		analysis, err := analyzeBootLogsFn(ctx, cfg, logs)
		if err == nil {
			return analysis, nil
		}
		lastErr = err
		if i < bootMaxRetries-1 {
			rich.Println("[yellow]Boot diagnosis attempt %d failed: %v. Retrying in %v...[/yellow]", i+1, err, bootRetryInterval)
			bootRetryDelay(bootRetryInterval)
		}
	}
	return "", fmt.Errorf("부팅 진단 %d회 시도 후 실패: %w", bootMaxRetries, lastErr)
}
