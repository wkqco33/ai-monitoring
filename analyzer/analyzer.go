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

func formatPrompt(state *monitor.SystemState) string {
	prompt := fmt.Sprintf("현재 시스템 상태:\n- CPU 사용량: %.2f%%\n- 메모리 사용량: %.2f%%\n\n상위 프로세스:\n", state.CPUUsage, state.MemUsage)
	for _, p := range state.Processes {
		prompt += fmt.Sprintf("PID: %d | 이름: %s | CPU: %.2f%% | Mem: %.2f%%\n", p.PID, p.Name, p.CPUProg, p.MemProg)
	}
	return prompt
}
