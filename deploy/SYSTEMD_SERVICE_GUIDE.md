# pcam systemd 서비스 등록 가이드

이 문서는 Ubuntu 시스템에서 `pcam`을 백그라운드 서비스(Daemon)로 등록하여 시스템 부팅 시 자동으로 실행되고 상태를 감시하도록 설정하는 방법을 안내합니다.

> **참고**: `service install` 커맨드를 사용하면 아래 과정을 자동으로 수행할 수 있습니다.
>
> ```bash
> sudo task install
> sudo ./pcam service install
> ```
>
> 유닛 파일 생성, 환경변수 템플릿 생성, 로그 파일 권한 설정, 서비스 활성화가 한 번에 처리됩니다.
> 수동으로 세밀하게 제어하려면 아래 과정을 따르세요.

## 1. 전제 조건

- `go` 컴파일러가 설치되어 있어야 합니다.
- `task` 도구가 설치되어 있어야 합니다. (https://taskfile.dev)
- Azure OpenAI API 키와 엔드포인트 정보가 필요합니다.

## 2. 기존 `ai-monitoring` 설치에서 마이그레이션

바이너리명이 `ai-monitoring`에서 `pcam`으로 변경되었습니다. 서비스명, 설정 경로, 환경변수 접두사(`AI_MONITORING_` → `PCAM_`)도 함께 바뀌었으므로 아래 순서로 이전 설치를 정리하세요.

```bash
# 1. 이전 서비스 중지 및 제거
sudo systemctl disable --now ai-monitoring
sudo rm -f /etc/systemd/system/ai-monitoring.service
sudo systemctl daemon-reload

# 2. 이전 바이너리 제거
sudo rm -f /usr/local/bin/ai-monitoring

# 3. 환경변수 파일 이관 (이름과 접두사 변경)
sudo mv /etc/default/ai-monitoring /etc/default/pcam
sudo sed -i 's/AI_MONITORING_/PCAM_/g' /etc/default/pcam

# 4. 설정 파일과 로그 이관 (선택)
mv ~/.config/ai-monitoring ~/.config/pcam
mkdir -p ~/.local/state/pcam
mv ~/.local/state/ai-monitoring/ai-monitoring.log ~/.local/state/pcam/pcam.log

# 5. 새 바이너리 설치 및 서비스 등록
sudo task install
sudo ./pcam service install
```

## 3. 설치 단계

### 1단계: 바이너리 빌드 및 설치

프로젝트 루트 디렉토리에서 다음 명령어를 실행하여 바이너리를 빌드하고 시스템 경로(`/usr/local/bin`)에 설치합니다.

```bash
sudo task install
```

### 2단계: 환경 변수 설정

API 키와 같은 민감한 정보를 안전하게 관리하기 위해 환경 변수 파일을 생성합니다.

```bash
# 템플릿 복사
sudo cp deploy/pcam.env.example /etc/default/pcam

# 파일 편집 (PCAM_ 접두사 형식으로 입력)
sudo nano /etc/default/pcam
```

`/etc/default/pcam` 파일 예시:

```env
PCAM_CHECK_INTERVAL=10s
PCAM_CPU_THRESHOLD=90.0
PCAM_MEMORY_THRESHOLD=90.0
PCAM_AZURE_ENDPOINT=https://your-resource.openai.azure.com/
PCAM_AZURE_API_KEY=your-api-key-here
PCAM_AZURE_DEPLOYMENT=gpt-4o
PCAM_BOT_TOKEN=your-bot-token-here
PCAM_COOLDOWN_PERIOD=5m
```

`pcam`는 `config/config.go`의 설정 규칙에 따라 `PCAM_` 접두사가 붙은 환경 변수만 읽습니다.
따라서 `AZURE_ENDPOINT` 같은 이름은 인식되지 않고, 반드시 `PCAM_AZURE_ENDPOINT` 형식으로 지정해야 합니다.

### 3단계: 로그 파일 생성 및 권한 설정

서비스가 로그를 기록할 수 있도록 파일을 생성하고 권한을 부여합니다. (현재 사용자: `seoyc` 기준)

```bash
sudo touch /var/log/pcam.log
sudo chown seoyc:seoyc /var/log/pcam.log
```

### 4단계: 서비스 파일 등록

작성된 `.service` 파일을 systemd 설정 디렉토리에 복사합니다.

```bash
sudo cp deploy/pcam.service /etc/systemd/system/
```

### 5단계: 서비스 시작 및 활성화

설정을 반영하고 서비스를 시작하며, 시스템 부팅 시 자동 실행되도록 등록합니다.

```bash
# systemd 설정 재로드
sudo systemctl daemon-reload

# 서비스 시작
sudo systemctl start pcam

# 부팅 시 자동 시작 등록
sudo systemctl enable pcam
```

## 4. 서비스 관리 명령어

| 작업                 | 명령어                        |
| :------------------- | :---------------------------- |
| **상태 확인**        | `sudo systemctl status pcam`  |
| **로그 실시간 확인** | `tail -f /var/log/pcam.log`   |
| **서비스 중지**      | `sudo systemctl stop pcam`    |
| **서비스 재시작**    | `sudo systemctl restart pcam` |
| **서비스 비활성화**  | `sudo systemctl disable pcam` |

## 5. 트러블슈팅

### 부팅 로그 접근 권한 문제

`monitor/boot.go`에서 사용하는 `journalctl -b` 명령어는 권한이 필요할 수 있습니다. 서비스가 실행 중인 사용자(`seoyc`)가 부팅 로그를 읽을 수 있도록 `systemd-journal` 그룹에 추가해야 합니다.

```bash
sudo usermod -aG systemd-journal seoyc
# 설정 적용을 위해 로그아웃 후 다시 로그인하거나 시스템을 재부팅하세요.
```

### API 호출 실패

`/var/log/pcam.log`를 확인하여 `Azure OpenAI 설정이 누락되었습니다` 또는 `Network Error`가 발생하는지 확인하세요. `/etc/default/pcam` 파일의 변수명이 `PCAM_` 접두사 형식인지 다시 점검하십시오.
