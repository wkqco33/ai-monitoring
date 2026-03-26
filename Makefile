# AI PC Monitoring System Makefile

BINARY_NAME=ai-monitoring
PREFIX=/usr/local
BINDIR=$(PREFIX)/bin

.PHONY: all build clean install uninstall test fmt vet tidy help

# 기본 타겟: 빌드
all: build

# 프로젝트 빌드
build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) main.go

# 빌드 결과물 삭제
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)

# 시스템에 설치 (sudo 권한이 필요할 수 있습니다)
install: build
	@echo "Installing $(BINARY_NAME) to $(BINDIR)..."
	install -d $(BINDIR)
	install -m 755 $(BINARY_NAME) $(BINDIR)/$(BINARY_NAME)

# 시스템에서 삭제
uninstall:
	@echo "Uninstalling $(BINARY_NAME) from $(BINDIR)..."
	rm -f $(BINDIR)/$(BINARY_NAME)

# 테스트 실행
test:
	@echo "Running tests..."
	go test -v ./...

# 코드 포맷팅
fmt:
	@echo "Formatting code..."
	go fmt ./...

# 정적 분석
vet:
	@echo "Running go vet..."
	go vet ./...

# 의존성 정리
tidy:
	@echo "Tidying go modules..."
	go mod tidy

# 도움말
help:
	@echo "사용 가능한 명령어:"
	@echo "  make build      - 프로젝트 빌드"
	@echo "  make clean      - 빌드 결과물 삭제"
	@echo "  make install    - $(BINDIR)에 설치 (sudo 권한 필요할 수 있음)"
	@echo "  make uninstall  - 설치된 바이너리 삭제"
	@echo "  make test       - 전체 테스트 실행"
	@echo "  make fmt        - 코드 포맷팅 적용"
	@echo "  make vet        - 정적 분석 실행"
	@echo "  make tidy       - 의존성(go.mod) 정리"
