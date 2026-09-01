package service

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
)

// ResolveInstallOptions 현재 실행 환경에서 설치 옵션을 구성합니다.
// sudo로 실행된 경우 서비스 실행 사용자는 SUDO_USER 기준으로 결정합니다.
func ResolveInstallOptions() (UnitOptions, error) {
	bin, err := os.Executable()
	if err != nil {
		return UnitOptions{}, fmt.Errorf("실행 파일 경로 확인 실패: %w", err)
	}
	binPath, err := filepath.Abs(bin)
	if err != nil {
		return UnitOptions{}, fmt.Errorf("실행 파일 경로 확인 실패: %w", err)
	}

	u, err := resolveRunUser(os.Getenv("SUDO_USER"))
	if err != nil {
		return UnitOptions{}, fmt.Errorf("서비스 실행 사용자 확인 실패: %w", err)
	}
	g, err := user.LookupGroupId(u.Gid)
	if err != nil {
		return UnitOptions{}, fmt.Errorf("서비스 실행 그룹 확인 실패: %w", err)
	}
	uid, _ := strconv.Atoi(u.Uid)
	gid, _ := strconv.Atoi(u.Gid)

	return UnitOptions{
		Description: "AI PC Monitoring System",
		BinaryPath:  binPath,
		UserName:    u.Username,
		GroupName:   g.Name,
		Uid:         uid,
		Gid:         gid,
		EnvFile:     DefaultEnvFile,
		LogFile:     DefaultLogFile,
	}, nil
}

// resolveRunUser 서비스를 실행할 사용자를 결정합니다.
func resolveRunUser(sudoUser string) (*user.User, error) {
	if sudoUser != "" && sudoUser != "root" {
		return user.Lookup(sudoUser)
	}
	return user.Current()
}
