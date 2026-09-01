// Package service는 systemd 서비스 등록과 제어를 담당합니다.
package service

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

const (
	DefaultServiceName = "ai-monitoring"
	DefaultUnitPath    = "/etc/systemd/system/ai-monitoring.service"
	DefaultEnvFile     = "/etc/default/ai-monitoring"
	DefaultLogFile     = "/var/log/ai-monitoring.log"
	DefaultBinaryPath  = "/usr/local/bin/" + DefaultServiceName
)

// DefaultEnvFileContent는 환경변수 파일이 없을 때 생성되는 템플릿입니다.
const DefaultEnvFileContent = `# ai-monitoring systemd service environment file
# AI_MONITORING_ 접두사 형식으로 작성해야 인식됩니다.

AI_MONITORING_CHECK_INTERVAL=10s
AI_MONITORING_CPU_THRESHOLD=90.0
AI_MONITORING_MEMORY_THRESHOLD=90.0

# Azure OpenAI
AI_MONITORING_AZURE_ENDPOINT=https://your-resource-name.openai.azure.com/
AI_MONITORING_AZURE_API_KEY=your-api-key-here
AI_MONITORING_AZURE_DEPLOYMENT=gpt-4o

# Ollama (llm_provider=ollama 일 때 사용)
# AI_MONITORING_LLM_PROVIDER=ollama
# AI_MONITORING_OLLAMA_ENDPOINT=http://localhost:11434
# AI_MONITORING_OLLAMA_MODEL=llama3

AI_MONITORING_BOT_TOKEN=
AI_MONITORING_COOLDOWN_PERIOD=5m
`

// Manager는 systemd 서비스 설치 및 제어를 담당합니다.
type Manager struct {
	ServiceName string
	UnitPath    string
	EnvFile     string
	LogFile     string

	// 외부 의존성. 테스트에서 교체할 수 있도록 필드로 유지합니다.
	run          func(name string, args ...string) (string, error)
	geteuid      func() int
	checkSystemd func() error
}

// New 기본 경로와 실제 명령 실행으로 Manager를 구성합니다.
func New() *Manager {
	return &Manager{
		ServiceName:  DefaultServiceName,
		UnitPath:     DefaultUnitPath,
		EnvFile:      DefaultEnvFile,
		LogFile:      DefaultLogFile,
		run:          runCmd,
		geteuid:      os.Geteuid,
		checkSystemd: checkSystemd,
	}
}

// runCmd 명령을 실행하고 출력과 에러를 반환합니다.
func runCmd(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).CombinedOutput()
	return string(out), err
}

// checkSystemd systemd 환경 사용 가능 여부를 확인합니다.
func checkSystemd() error {
	if runtime.GOOS != "linux" {
		return errors.New("systemd 서비스는 Linux에서만 지원됩니다")
	}
	if _, err := exec.LookPath("systemctl"); err != nil {
		return errors.New("systemctl을 찾을 수 없습니다: systemd 환경이 아닌 것 같습니다")
	}
	return nil
}

// requireRoot 시스템 파일을 수정할 수 있는 권한이 있는지 확인합니다.
func (m *Manager) requireRoot() error {
	if m.geteuid() != 0 {
		return errors.New("root 권한이 필요합니다: sudo ai-monitoring service ...")
	}
	return nil
}

// Install 서비스 유닛 파일을 등록하고 서비스를 활성화·시작합니다.
func (m *Manager) Install(opts UnitOptions) error {
	if err := m.requireRoot(); err != nil {
		return err
	}
	if err := m.checkSystemd(); err != nil {
		return err
	}

	if err := m.writeUnitFile(RenderUnit(opts)); err != nil {
		return err
	}
	if err := m.ensureEnvFile(); err != nil {
		return err
	}
	if err := ensureLogFile(m.LogFile, opts.Uid, opts.Gid); err != nil {
		return fmt.Errorf("로그 파일 생성 실패: %w", err)
	}

	if _, err := m.run("systemctl", "daemon-reload"); err != nil {
		return fmt.Errorf("daemon-reload 실패: %v", err)
	}
	if _, err := m.run("systemctl", "enable", "--now", m.ServiceName); err != nil {
		return fmt.Errorf("서비스 활성화 실패: %v", err)
	}
	return nil
}

// writeUnitFile 유닛 파일을 작성합니다. 기존 내용이 다르면 백업을 남깁니다.
func (m *Manager) writeUnitFile(content string) error {
	if existing, err := os.ReadFile(m.UnitPath); err == nil && string(existing) != content {
		if err := os.WriteFile(m.UnitPath+".bak", existing, 0o644); err != nil {
			return fmt.Errorf("기존 유닛 파일 백업 실패: %w", err)
		}
	}
	if err := os.WriteFile(m.UnitPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("유닛 파일 작성 실패: %w", err)
	}
	return nil
}

// ensureEnvFile 환경변수 파일이 없으면 템플릿으로 생성합니다. 기존 파일은 절대 덮어쓰지 않습니다.
func (m *Manager) ensureEnvFile() error {
	if _, err := os.Stat(m.EnvFile); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.WriteFile(m.EnvFile, []byte(DefaultEnvFileContent), 0o600); err != nil {
		return fmt.Errorf("환경변수 파일 생성 실패: %w", err)
	}
	return nil
}

// ensureLogFile 로그 파일이 없으면 생성하고 서비스 실행 사용자에게 소유권을 부여합니다.
// uid/gid가 -1이면 소유권을 변경하지 않습니다.
func ensureLogFile(path string, uid, gid int) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	f.Close()
	return os.Chown(path, uid, gid)
}

// Uninstall 서비스를 비활성화하고 유닛 파일을 제거합니다.
func (m *Manager) Uninstall() error {
	if err := m.requireRoot(); err != nil {
		return err
	}
	if err := m.checkSystemd(); err != nil {
		return err
	}

	// 등록되지 않은 서비스도 정리를 계속할 수 있어야 하므로 실패를 무시합니다.
	_, _ = m.run("systemctl", "disable", "--now", m.ServiceName)

	if err := os.Remove(m.UnitPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("유닛 파일 삭제 실패: %w", err)
	}
	if _, err := m.run("systemctl", "daemon-reload"); err != nil {
		return fmt.Errorf("daemon-reload 실패: %v", err)
	}
	return nil
}

// Status 서비스 상태를 조회합니다. 서비스가 비활성(exit 3)인 것은 정상 상태입니다.
func (m *Manager) Status() (string, error) {
	if err := m.checkSystemd(); err != nil {
		return "", err
	}
	out, err := m.run("systemctl", "status", m.ServiceName, "--no-pager")
	var exitErr *exec.ExitError
	if err != nil && !errors.As(err, &exitErr) {
		return out, err
	}
	return out, nil
}

// Start 서비스를 시작합니다.
func (m *Manager) Start() error { return m.action("start") }

// Stop 서비스를 중지합니다.
func (m *Manager) Stop() error { return m.action("stop") }

// Restart 서비스를 재시작합니다.
func (m *Manager) Restart() error { return m.action("restart") }

// action systemctl 동사를 실행하고 실패 시 출력을 포함한 에러를 반환합니다.
func (m *Manager) action(verb string) error {
	if err := m.checkSystemd(); err != nil {
		return err
	}
	out, err := m.run("systemctl", verb, m.ServiceName)
	if err != nil {
		return fmt.Errorf("systemctl %s 실패: %v\n%s\nroot 권한이 필요할 수 있습니다: sudo ai-monitoring service %s", verb, err, out, verb)
	}
	return nil
}
