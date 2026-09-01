package analyzer

import (
	"context"
	"strings"
	"testing"

	"ai-monitoring/config"
	"ai-monitoring/monitor"
	llm "github.com/wkqco33/LLM_client_go"
)

// mockClient는 LLM 클라이언트를 대체하는 목 구현체입니다.
type mockClient struct {
	resp *llm.ChatResponse
	err  error
	req  llm.ChatRequest
}

func (m *mockClient) Complete(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
	m.req = req
	return m.resp, m.err
}

// withMockClient는 테스트 동안 LLM 클라이언트 생성을 목으로 교체합니다.
func withMockClient(t *testing.T, mc *mockClient) {
	t.Helper()
	orig := newClient
	newClient = func(cfg *config.AppConfig) (LLMClient, error) {
		return mc, nil
	}
	t.Cleanup(func() { newClient = orig })
}

func testConfig() *config.AppConfig {
	return &config.AppConfig{
		LLMProvider:     "azure",
		AzureEndpoint:   "https://example.openai.azure.com/",
		AzureOpenAIKey:  "test-key",
		AzureDeployment: "gpt-4o",
	}
}

func TestAnalyzeSystemStateWithMockClient(t *testing.T) {
	mc := &mockClient{resp: &llm.ChatResponse{
		Choices: []llm.Choice{{Message: llm.Message{Content: "분석 완료"}}},
	}}
	withMockClient(t, mc)

	state := &monitor.SystemState{CPUUsage: 95.5, MemUsage: 80.2}
	got, err := AnalyzeSystemState(context.Background(), testConfig(), state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "분석 완료" {
		t.Errorf("expected 분석 완료, got %q", got)
	}
	if len(mc.req.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(mc.req.Messages))
	}
	if mc.req.Messages[0].Role != llm.RoleSystem || mc.req.Messages[1].Role != llm.RoleUser {
		t.Errorf("unexpected roles: %s, %s", mc.req.Messages[0].Role, mc.req.Messages[1].Role)
	}
	if mc.req.Model != "gpt-4o" {
		t.Errorf("expected model gpt-4o, got %q", mc.req.Model)
	}
}

func TestAnalyzeBootLogsWithMockClient(t *testing.T) {
	mc := &mockClient{resp: &llm.ChatResponse{
		Choices: []llm.Choice{{Message: llm.Message{Content: "부팅 이상 없음"}}},
	}}
	withMockClient(t, mc)

	got, err := AnalyzeBootLogs(context.Background(), testConfig(), "kernel panic at boot")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "부팅 이상 없음" {
		t.Errorf("expected 부팅 이상 없음, got %q", got)
	}
	if !strings.Contains(mc.req.Messages[1].Content, "kernel panic") {
		t.Errorf("user prompt should contain logs, got %q", mc.req.Messages[1].Content)
	}
}

func TestFormatPrompt(t *testing.T) {
	state := &monitor.SystemState{
		CPUUsage: 95.5,
		MemUsage: 80.2,
		Processes: []monitor.ProcessInfo{
			{PID: 1234, Name: "chrome", CPUProg: 40.0, MemProg: 15.0},
			{PID: 5678, Name: "vscode", CPUProg: 30.0, MemProg: 10.0},
		},
	}

	prompt := formatPrompt(state)

	if !strings.Contains(prompt, "95.50%") {
		t.Errorf("Expected prompt to contain CPU usage, got: %s", prompt)
	}
	if !strings.Contains(prompt, "chrome") {
		t.Errorf("Expected prompt to contain chrome, got: %s", prompt)
	}
}
