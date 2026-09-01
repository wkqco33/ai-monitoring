package analyzer

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"ai-monitoring/config"
	"ai-monitoring/monitor"
	llm "github.com/wkqco33/LLM_client_go"
	"github.com/wkqco33/LLM_client_go/azure"
	"github.com/wkqco33/LLM_client_go/ollama"
)

// LLMClient는 다양한 LLM 제공자를 추상화하는 인터페이스입니다.
type LLMClient interface {
	Complete(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error)
}

// newClient는 테스트에서 목 클라이언트로 교체할 수 있도록 패키지 변수로 유지합니다.
var newClient = getClient

// normalizeOllamaEndpoint는 Ollama의 OpenAI 호환 API 경로(/v1)를 보장합니다.
// 사용자가 호스트만 입력해도 /v1이 자동으로 붙도록 하여 잘못된 경로로 요청이 가는 것을 방지합니다.
func normalizeOllamaEndpoint(endpoint string) string {
	if endpoint == "" || strings.HasSuffix(endpoint, "/v1") || strings.HasSuffix(endpoint, "/v1/") {
		return endpoint
	}
	return strings.TrimRight(endpoint, "/") + "/v1"
}

// getClient는 설정에 따라 적절한 LLM 클라이언트를 반환합니다.
func getClient(cfg *config.AppConfig) (LLMClient, error) {
	switch cfg.LLMProvider {
	case "ollama":
		return ollama.New(ollama.Config{
			BaseURL: normalizeOllamaEndpoint(cfg.OllamaEndpoint),
		}), nil
	case "azure":
		if cfg.AzureEndpoint == "" || cfg.AzureOpenAIKey == "" {
			return nil, fmt.Errorf("Azure OpenAI 설정이 누락되었습니다")
		}
		return azure.New(azure.Config{
			Endpoint: cfg.AzureEndpoint,
			APIKey:   cfg.AzureOpenAIKey,
		}), nil
	default:
		return nil, fmt.Errorf("지원하지 않는 LLM 제공자입니다: %s", cfg.LLMProvider)
	}
}

// complete는 시스템/사용자 프롬프트로 LLM 완성 요청을 보내고 응답 텍스트를 반환합니다.
func complete(ctx context.Context, cfg *config.AppConfig, systemPrompt, userPrompt string) (string, error) {
	client, err := newClient(cfg)
	if err != nil {
		return "", err
	}

	model := cfg.AzureDeployment
	if cfg.LLMProvider == "ollama" {
		model = cfg.OllamaModel
	}

	req := llm.ChatRequest{
		Model: model,
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: systemPrompt},
			{Role: llm.RoleUser, Content: userPrompt},
		},
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, 90*time.Second)
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

// AnalyzeSystemState 시스템 상태를 분석합니다.
func AnalyzeSystemState(ctx context.Context, cfg *config.AppConfig, state *monitor.SystemState) (string, error) {
	prompt := formatPrompt(state)
	slog.Debug("LLM에 프롬프트 전송", "prompt", prompt)

	systemPrompt := "당신은 시스템 모니터링 전문가입니다. 제공된 시스템 상태를 분석하고, 리소스 사용량이 높은 원인과 해결 방안을 한국어로 간결하게 제시해주세요."
	return complete(ctx, cfg, systemPrompt, prompt)
}

// AnalyzeBootLogs 부팅 로그를 진단합니다.
func AnalyzeBootLogs(ctx context.Context, cfg *config.AppConfig, logs string) (string, error) {
	userPrompt := fmt.Sprintf("다음은 시스템 부팅 시 발생한 주요 로그 또는 에러 메시지입니다:\n\n%s\n\n이 로그들을 분석하여 시스템 안정성에 문제가 없는지 진단하고, 발견된 특이사항와 조치 방법을 한국어로 간결하게 설명해주세요.", logs)

	systemPrompt := "당신은 시스템 보안 및 하드웨어 전문가입니다. 부팅 로그를 분석하여 잠재적인 하드웨어 문제나 소프트웨어 충돌을 진단합니다."
	return complete(ctx, cfg, systemPrompt, userPrompt)
}

// maxLogPayloadChars는 LLM 요청에 담을 로그 페이로드의 최대 길이(문자 수)입니다.
const maxLogPayloadChars = 8000

// AnalyzeRecentLogs 최근 애플리케이션 로그를 분석해 특이사항을 요약합니다.
func AnalyzeRecentLogs(ctx context.Context, cfg *config.AppConfig, logs string) (string, error) {
	if strings.TrimSpace(logs) == "" {
		return "", fmt.Errorf("분석할 로그가 없습니다")
	}

	userPrompt := fmt.Sprintf("다음은 pcam의 최근 동작 로그입니다:\n\n%s\n\n이 로그를 분석하여 반복되는 에러, 비정상 패턴, 임계치 초과 이력 등 주요 특이사항과 조치 방법을 한국어로 간결하게 정리해주세요.", truncateTail(logs, maxLogPayloadChars))

	systemPrompt := "당신은 시스템 모니터링 로그 분석 전문가입니다. 주어진 로그에서 이상 징후를 찾아 원인과 조치 방안을 진단합니다."
	return complete(ctx, cfg, systemPrompt, userPrompt)
}

// truncateTail은 문자열이 max를 초과하면 앞부분을 잘라내고 최근(뒷부분) 로그를 유지합니다.
// 반환값은 마커를 포함해 max 문자를 넘지 않습니다.
func truncateTail(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	const marker = "...(앞부분 생략)...\n"
	keep := max - len([]rune(marker))
	if keep < 0 {
		keep = 0
	}
	return marker + string(runes[len(runes)-keep:])
}

func formatPrompt(state *monitor.SystemState) string {
	prompt := fmt.Sprintf("현재 시스템 상태:\n- CPU 사용량: %.2f%%\n- 메모리 사용량: %.2f%%\n\n상위 프로세스:\n", state.CPUUsage, state.MemUsage)
	for _, p := range state.Processes {
		prompt += fmt.Sprintf("PID: %d | 이름: %s | CPU: %.2f%% | Mem: %.2f%%\n", p.PID, p.Name, p.CPUProg, p.MemProg)
	}
	return prompt
}
