package service

import "fmt"

// UnitOptions는 systemd 유닛 파일 생성에 필요한 정보입니다.
type UnitOptions struct {
	Description string
	BinaryPath  string
	UserName    string
	GroupName   string
	EnvFile     string
	LogFile     string
}

// RenderUnit systemd 유닛 파일 내용을 생성합니다.
func RenderUnit(opts UnitOptions) string {
	return fmt.Sprintf(`[Unit]
Description=%s
After=network-online.target syslog.target
Wants=network-online.target

[Service]
Type=simple
User=%s
Group=%s
EnvironmentFile=-%s
ExecStart=%s start
Restart=always
RestartSec=5

StandardOutput=append:%s
StandardError=append:%s

[Install]
WantedBy=multi-user.target
`,
		opts.Description,
		opts.UserName,
		opts.GroupName,
		opts.EnvFile,
		opts.BinaryPath,
		opts.LogFile,
		opts.LogFile,
	)
}
