# ai-monitoring systemd 서비스 등록 가이드

이 문서는 Ubuntu 시스템에서 `ai-monitoring`을 백그라운드 서비스(Daemon)로 등록하여 시스템 부팅 시 자동으로 실행되고 상태를 감시하도록 설정하는 방법을 안내합니다.

## 1. 전제 조건

- `go` 컴파일러가 설치되어 있어야 합니다.
- `make` 도구가 설치되어 있어야 합니다.
- Azure OpenAI API 키와 엔드포인트 정보가 필요합니다.

## 2. 설치 단계

### 1단계: 바이너리 빌드 및 설치
프로젝트 루트 디렉토리에서 다음 명령어를 실행하여 바이너리를 빌드하고 시스템 경로(`/usr/local/bin`)에 설치합니다.

```bash
make install
```

### 2단계: 환경 변수 설정
API 키와 같은 민감한 정보를 안전하게 관리하기 위해 환경 변수 파일을 생성합니다.

```bash
# 템플릿 복사
sudo cp ai-monitoring.env.example /etc/default/ai-monitoring

# 파일 편집 (AI_MONITORING_ 접두사 형식으로 입력)
sudo nano /etc/default/ai-monitoring
```

`/etc/default/ai-monitoring` 파일 예시:
```env
AI_MONITORING_CHECK_INTERVAL=10s
AI_MONITORING_CPU_THRESHOLD=90.0
AI_MONITORING_MEMORY_THRESHOLD=90.0
AI_MONITORING_AZURE_ENDPOINT=https://your-resource.openai.azure.com/
AI_MONITORING_AZURE_API_KEY=your-api-key-here
AI_MONITORING_AZURE_DEPLOYMENT=gpt-4o
AI_MONITORING_BOT_TOKEN=your-bot-token-here
AI_MONITORING_COOLDOWN_PERIOD=5m
```

`ai-monitoring`는 `config/config.go`의 설정 규칙에 따라 `AI_MONITORING_` 접두사가 붙은 환경 변수만 읽습니다.
따라서 `AZURE_ENDPOINT` 같은 이름은 인식되지 않고, 반드시 `AI_MONITORING_AZURE_ENDPOINT` 형식으로 지정해야 합니다.

### 3단계: 로그 파일 생성 및 권한 설정
서비스가 로그를 기록할 수 있도록 파일을 생성하고 권한을 부여합니다. (현재 사용자: `seoyc` 기준)

```bash
sudo touch /var/log/ai-monitoring.log
sudo chown seoyc:seoyc /var/log/ai-monitoring.log
```

### 4단계: 서비스 파일 등록
작성된 `.service` 파일을 systemd 설정 디렉토리에 복사합니다.

```bash
sudo cp ai-monitoring.service /etc/systemd/system/
```

### 5단계: 서비스 시작 및 활성화
설정을 반영하고 서비스를 시작하며, 시스템 부팅 시 자동 실행되도록 등록합니다.

```bash
# systemd 설정 재로드
sudo systemctl daemon-reload

# 서비스 시작
sudo systemctl start ai-monitoring

# 부팅 시 자동 시작 등록
sudo systemctl enable ai-monitoring
```

## 3. 서비스 관리 명령어

| 작업 | 명령어 |
| :--- | :--- |
| **상태 확인** | `sudo systemctl status ai-monitoring` |
| **로그 실시간 확인** | `tail -f /var/log/ai-monitoring.log` |
| **서비스 중지** | `sudo systemctl stop ai-monitoring` |
| **서비스 재시작** | `sudo systemctl restart ai-monitoring` |
| **서비스 비활성화** | `sudo systemctl disable ai-monitoring` |

## 4. 트러블슈팅

### 부팅 로그 접근 권한 문제
`monitor/boot.go`에서 사용하는 `journalctl -b` 명령어는 권한이 필요할 수 있습니다. 서비스가 실행 중인 사용자(`seoyc`)가 부팅 로그를 읽을 수 있도록 `systemd-journal` 그룹에 추가해야 합니다.

```bash
sudo usermod -aG systemd-journal seoyc
# 설정 적용을 위해 로그아웃 후 다시 로그인하거나 시스템을 재부팅하세요.
```

### API 호출 실패
`/var/log/ai-monitoring.log`를 확인하여 `Azure OpenAI 설정이 누락되었습니다` 또는 `Network Error`가 발생하는지 확인하세요. `/etc/default/ai-monitoring` 파일의 변수명이 `AI_MONITORING_` 접두사 형식인지 다시 점검하십시오.
