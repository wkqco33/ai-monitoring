package analyzer

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	llm "llm-client-go"
	"llm-client-go/azure"
	"ai-monitoring/config"
	"ai-monitoring/monitor"
)

// AnalyzeSystemState Azure OpenAI를 사용하여 시스템 상태를 분석합니다.
func AnalyzeSystemState(ctx context.Context, cfg *config.AppConfig, state *monitor.SystemState) (string, error) {
	if cfg.AzureEndpoint == "" || cfg.AzureOpenAIKey == "" {
		return "", fmt.Errorf("Azure OpenAI 설정이 누락되었습니다")
	}

	client := azure.New(azure.Config{
		Endpoint: cfg.AzureEndpoint,
		APIKey:   cfg.AzureOpenAIKey,
	})

	prompt := formatPrompt(state)
	slog.Debug("LLM에 프롬프트 전송", "prompt", prompt)

	req := llm.ChatRequest{
		Model: cfg.AzureDeployment,
		Messages: []llm.Message{
			{
				Role:    llm.RoleSystem,
				Content: "당신은 시스템 모니터링 전문가입니다. 제공된 시스템 상태를 분석하고, 리소스 사용량이 높은 원인과 해결 방안을 한국어로 간결하게 제시해주세요.",
			},
			{
				Role:    llm.RoleUser,
				Content: prompt,
			},
		},
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	resp, err := client.Complete(ctxTimeout, req)
	if err != nil {
		return "", fmt.Errorf("LLM 분석 요청 실패: %w", err)
	}

	if len(resp.Choices) > 0 {
		return resp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("LLM에서 응답을 받지 못했습니다")
}

// AnalyzeBootLogs Azure OpenAI를 사용하여 부팅 로그를 진단합니다.
func AnalyzeBootLogs(ctx context.Context, cfg *config.AppConfig, logs string) (string, error) {
	if cfg.AzureEndpoint == "" || cfg.AzureOpenAIKey == "" {
		return "", fmt.Errorf("Azure OpenAI 설정이 누락되었습니다")
	}

	client := azure.New(azure.Config{
		Endpoint: cfg.AzureEndpoint,
		APIKey:   cfg.AzureOpenAIKey,
	})

	prompt := fmt.Sprintf("다음은 시스템 부팅 시 발생한 주요 로그 또는 에러 메시지입니다:\n\n%s\n\n이 로그들을 분석하여 시스템 안정성에 문제가 없는지 진단하고, 발견된 특이사항과 조치 방법을 한국어로 간결하게 설명해주세요.", logs)

	req := llm.ChatRequest{
		Model: cfg.AzureDeployment,
		Messages: []llm.Message{
			{
				Role:    llm.RoleSystem,
				Content: "당신은 시스템 보안 및 하드웨어 전문가입니다. 부팅 로그를 분석하여 잠재적인 하드웨어 문제나 소프트웨어 충돌을 진단합니다.",
			},
			{
				Role:    llm.RoleUser,
				Content: prompt,
			},
		},
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	resp, err := client.Complete(ctxTimeout, req)
	if err != nil {
		return "", fmt.Errorf("부팅 진단 LLM 요청 실패: %w", err)
	}

	if len(resp.Choices) > 0 {
		return resp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("부팅 진단 응답을 받지 못했습니다")
}

func formatPrompt(state *monitor.SystemState) string {
	prompt := fmt.Sprintf("현재 시스템 상태:\n- CPU 사용량: %.2f%%\n- 메모리 사용량: %.2f%%\n\n상위 프로세스:\n", state.CPUUsage, state.MemUsage)
	for _, p := range state.Processes {
		prompt += fmt.Sprintf("PID: %d | 이름: %s | CPU: %.2f%% | Mem: %.2f%%\n", p.PID, p.Name, p.CPUProg, p.MemProg)
	}
	return prompt
}
