package notifier

import (
	"log/slog"

	"github.com/gen2brain/beeep"
	"ai-monitoring/config"
)

// Notify 분석 결과를 사용자에게 알립니다.
func Notify(cfg *config.AppConfig, title, message string) {
	// OS 알림
	err := beeep.Notify(title, message, "")
	if err != nil {
		slog.Error("OS 알림 전송 실패", "error", err)
	} else {
		slog.Info("OS 알림 전송 성공", "title", title)
	}

	// 봇 알림 (설정된 경우)
	if cfg.BotToken != "" {
		// TODO: llm-client-go/bots 패키지를 활용한 봇 전송 로직 구현
		slog.Info("봇 알림 전송 (TODO)", "token", cfg.BotToken)
	}
}
