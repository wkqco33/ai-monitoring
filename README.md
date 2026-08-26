# ai-monitoring

AI를 이용해 시스템 상태 이상을 감지하고 분석하는 PC 모니터링 도구입니다.

CPU와 메모리 사용률을 주기적으로 확인하고, 임계치를 넘으면 현재 프로세스 상태를 수집해 Azure OpenAI로 원인 분석을 요청합니다. 또한 부팅 로그를 함께 검사해 초기 이상 징후도 진단합니다.

## 주요 기능

- CPU 및 메모리 사용률 주기적 모니터링
- 상위 프로세스 정보 수집
- 임계치 초과 시 LLM 기반 이상 분석
- 운영체제별 부팅 로그 진단
- 데스크톱 OS 알림 전송
- systemd 서비스 등록 지원

## 동작 방식

1. `monitor` 패키지가 `CheckInterval` 주기로 시스템 상태를 수집합니다.
2. CPU 또는 메모리 사용률이 설정된 임계치를 넘으면 이상 상태로 판단합니다.
3. `analyzer` 패키지가 현재 상태와 상위 프로세스 정보를 Azure OpenAI로 전달해 요약 분석을 받습니다.
4. `notifier` 패키지가 OS 알림을 띄웁니다.
5. 시작 시점에는 부팅 로그도 확인해 최근 에러를 별도로 진단합니다.

## 프로젝트 구조

- `main.go`: 프로그램 진입점
- `cmd/`: CLI 명령 정의
- `config/`: 설정 로딩 및 기본 경로 처리
- `monitor/`: 시스템 상태 및 부팅 로그 수집
- `analyzer/`: Azure OpenAI 분석 요청
- `notifier/`: 알림 전송
- `logger/`: 로그 초기화
- `config.example.yaml`: 설정 예시
- `ai-monitoring.service`: systemd 서비스 파일 예시

## 요구 사항

- Go 1.26.1 이상
- `task` (Taskfile 러너) — 설치: https://taskfile.dev
- Azure OpenAI 사용 시 엔드포인트와 API 키
- Linux에서 부팅 로그 확인 시 `journalctl` 접근 권한
- OS 알림 사용 시 데스크톱 알림을 지원하는 환경

## 설치

### 소스 빌드

```bash
task build
```

빌드 결과물은 프로젝트 루트의 `ai-monitoring` 바이너리로 생성됩니다.

### 시스템 설치

```bash
sudo task install
```

기본 설치 경로는 `/usr/local/bin/ai-monitoring` 입니다.

## 설정

설정은 YAML 파일 또는 환경변수로 지정할 수 있습니다.

### 1. 설정 파일 사용

기본 설정 파일 경로는 다음과 같습니다.

```bash
~/.config/ai-monitoring/config.yaml
```

예시 파일는 `config.example.yaml`에 있습니다. 이 파일을 복사해 실제 값으로 수정한 뒤 사용하세요.

### 2. 환경변수 사용

환경변수는 `AI_MONITORING_` 접두사를 사용합니다. 예를 들어 `azure_endpoint`는 `AI_MONITORING_AZURE_ENDPOINT` 로 지정합니다.

예시:

```bash
export AI_MONITORING_LLM_PROVIDER=ollama # azure 또는 ollama
export AI_MONITORING_OLLAMA_ENDPOINT=http://localhost:11434
export AI_MONITORING_OLLAMA_MODEL=llama3
export AI_MONITORING_CHECK_INTERVAL=10s
export AI_MONITORING_CPU_THRESHOLD=90.0
export AI_MONITORING_MEMORY_THRESHOLD=90.0
export AI_MONITORING_AZURE_ENDPOINT=https://your-resource-name.openai.azure.com/
export AI_MONITORING_AZURE_API_KEY=your-api-key-here
export AI_MONITORING_AZURE_DEPLOYMENT=gpt-4o
export AI_MONITORING_BOT_TOKEN=your-bot-token-here
export AI_MONITORING_COOLDOWN_PERIOD=5m
```

### 설정 항목

| 키                 | 타입     | 기본값                   | 설명                           |
| ------------------ | -------- | ------------------------ | ------------------------------ |
| `llm_provider`     | string   | `azure`                  | LLM 제공자 (`azure`, `ollama`) |
| `ollama_endpoint`  | string   | `http://localhost:11434` | Ollama 서버 API 엔드포인트     |
| `ollama_model`     | string   | `llama3`                 | Ollama에서 사용할 모델명       |
| `check_interval`   | duration | `10s`                    | 시스템 상태를 확인하는 주기    |
| `cpu_threshold`    | float    | `90.0`                   | CPU 사용률 경고 기준(%)        |
| `memory_threshold` | float    | `90.0`                   | 메모리 사용률 경고 기준(%)     |
| `azure_endpoint`   | string   | 없음                     | Azure OpenAI 엔드포인트        |
| `azure_api_key`    | string   | 없음                     | Azure OpenAI API 키            |
| `azure_deployment` | string   | 없음                     | Azure OpenAI 배포 이름         |
| `bot_token`        | string   | 없음                     | 알림 봇 토큰 예약 값           |
| `cooldown_period`  | duration | `5m`                     | 동일 알림 재전송 최소 간격     |

### 예시 `config.yaml`

```yaml
check_interval: 10s
cpu_threshold: 90.0
memory_threshold: 90.0
azure_endpoint: https://your-resource-name.openai.azure.com/
azure_api_key: your-api-key-here
azure_deployment: gpt-4o
bot_token: your-bot-token-here
cooldown_period: 5m
```

## 실행

기본 설정 파일 경로를 사용하면 다음처럼 실행할 수 있습니다.

```bash
./ai-monitoring start
```

설정 파일 경로를 명시하려면 다음과 같이 실행합니다.

```bash
./ai-monitoring --config ~/.config/ai-monitoring/config.yaml start
```

디버그 로그를 켜려면 `--debug` 옵션을 추가합니다.

```bash
./ai-monitoring --debug start
```

## CLI 옵션

### 전역 옵션

- `--debug`, `-d`: 디버그 로그 활성화
- `--config`, `-c`: 설정 파일 경로 지정

### `start` 옵션

- `--interval`, `-i`: 점검 주기
- `--cpu`, `-c`: CPU 임계치
- `--mem`, `-m`: 메모리 임계치
- `--cooldown`, `-C`: 알림 쿨다운
- `--azure-endpoint`, `-e`: Azure OpenAI 엔드포인트
- `--azure-key`, `-k`: Azure OpenAI API 키
- `--azure-deployment`, `-D`: Azure OpenAI 배포 이름

CLI로 명시한 값은 설정 파일 값을 덮어씁니다.

## systemd 서비스로 실행

`ai-monitoring.service` 파일과 `SYSTEMD_SERVICE_GUIDE.md`를 참고해 systemd 서비스로 등록할 수 있습니다.

핵심 흐름은 다음과 같습니다.

1. 바이너리를 설치합니다.
2. `/etc/default/ai-monitoring` 에 Azure 관련 환경변수를 설정합니다.
3. `ai-monitoring.service` 를 `/etc/systemd/system/` 에 복사합니다.
4. `systemctl enable --now ai-monitoring` 로 서비스 등록 및 시작을 수행합니다.

## 알림 기능 상태

- OS 알림은 구현되어 있습니다.
- `bot_token` 은 설정 키로 존재하지만, 봇 전송 로직은 아직 TODO 상태입니다.

## 트러블슈팅

- 부팅 로그를 읽을 수 없다는 오류가 나면 `journalctl` 권한을 확인하세요.
- Azure 분석이 실패하면 `azure_endpoint`, `azure_api_key`, `azure_deployment` 값을 확인하세요.
- 알림이 뜨지 않으면 데스크톱 세션에서 실행 중인지 확인하세요.

## 개발 명령

```bash
task test
task fmt
task vet
task tidy
```

개발 워크플로우(TDD)와 코드 스타일은 [`AGENTS.md`](AGENTS.md)를, 기여 절차는 [`CONTRIBUTING.md`](CONTRIBUTING.md)를 참고하세요.

## 보안

보안 취약점 보고와 민감 정보 처리 정책은 [`SECURITY.md`](SECURITY.md)를 참고하세요.

## 라이선스

이 프로젝트는 [MIT 라이선스](LICENSE)로 배포됩니다.
