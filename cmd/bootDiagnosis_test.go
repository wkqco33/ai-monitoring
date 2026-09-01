package cmd

import (
	"context"
	"errors"
	"testing"
	"time"

	"ai-monitoring/config"
)

// stubBootDeps는 부팅 진단 헬퍼의 외부 의존성을 교체합니다.
func stubBootDeps(t *testing.T, logs string, logsErr error, analyses []error, results []string) *[]time.Duration {
	t.Helper()
	getBootLogsFn = func() (string, error) { return logs, logsErr }

	idx := 0
	analyzeBootLogsFn = func(ctx context.Context, cfg *config.AppConfig, logs string) (string, error) {
		var err error
		if idx < len(analyses) {
			err = analyses[idx]
		} else if len(analyses) > 0 {
			err = analyses[len(analyses)-1]
		}
		result := ""
		if idx < len(results) {
			result = results[idx]
		}
		idx++
		return result, err
	}

	sleeps := &[]time.Duration{}
	bootRetryDelay = func(d time.Duration) { *sleeps = append(*sleeps, d) }

	t.Cleanup(func() {
		getBootLogsFn = nil
		analyzeBootLogsFn = nil
		bootRetryDelay = nil
	})
	return sleeps
}

func TestRunBootDiagnosisSuccess(t *testing.T) {
	stubBootDeps(t, "kernel error log", nil, []error{nil}, []string{"진단 결과"})

	got, err := runBootDiagnosis(context.Background(), &config.AppConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "진단 결과" {
		t.Errorf("expected 진단 결과, got %q", got)
	}
}

func TestRunBootDiagnosisEmptyLogsSkipsLLM(t *testing.T) {
	sleeps := stubBootDeps(t, "  \n ", nil, []error{nil}, []string{"호출되면 안 됨"})

	got, err := runBootDiagnosis(context.Background(), &config.AppConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty result, got %q", got)
	}
	if len(*sleeps) != 0 {
		t.Error("no retry should happen for empty logs")
	}
}

func TestRunBootDiagnosisFetchError(t *testing.T) {
	stubBootDeps(t, "", errors.New("journalctl failed"), nil, nil)

	if _, err := runBootDiagnosis(context.Background(), &config.AppConfig{}); err == nil {
		t.Fatal("expected error when fetching logs fails")
	}
}

func TestRunBootDiagnosisRetriesThenSucceeds(t *testing.T) {
	sleeps := stubBootDeps(t, "some log", nil,
		[]error{errors.New("api error"), errors.New("api error"), nil},
		[]string{"", "", "재시도 성공"})

	got, err := runBootDiagnosis(context.Background(), &config.AppConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "재시도 성공" {
		t.Errorf("expected 재시도 성공, got %q", got)
	}
	if len(*sleeps) != 2 {
		t.Errorf("expected 2 retries, got %d", len(*sleeps))
	}
}

func TestRunBootDiagnosisAllRetriesFailed(t *testing.T) {
	sleeps := stubBootDeps(t, "some log", nil,
		[]error{errors.New("api error")},
		[]string{""})

	_, err := runBootDiagnosis(context.Background(), &config.AppConfig{})
	if err == nil {
		t.Fatal("expected error after all retries failed")
	}
	if len(*sleeps) != bootMaxRetries-1 {
		t.Errorf("expected %d retries, got %d", bootMaxRetries-1, len(*sleeps))
	}
}
