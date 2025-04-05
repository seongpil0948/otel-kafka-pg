.PHONY: build run test clean proto docker-build docker-compose-up docker-compose-down

# 기본 변수 설정
BINARY_NAME=telemetry-backend
GO_BUILD_ENV=CGO_ENABLED=1
GO_FILES=$(shell find . -type f -name "*.go" -not -path "./proto/*")
PROTO_OUT_DIR=./proto

# 컴파일 및 빌드
build:
	$(GO_BUILD_ENV) go build -o ./bin/$(BINARY_NAME) ./cmd/app/main.go
	$(GO_BUILD_ENV) go build -o ./bin/healthcheck ./cmd/healthcheck/main.go

# 실행
run:
	$(GO_BUILD_ENV) go run ./cmd/app/main.go

# 단위 테스트
test:
	go test -v ./...

# 정적 코드 분석
lint:
	golangci-lint run

# 빌드 결과물 삭제
clean:
	rm -rf ./bin/*
	rm -rf $(PROTO_OUT_DIR)/*

# protoc 설치 확인 및 의존성 설치
check-protoc:
	@which protoc > /dev/null || (echo "protoc가 설치되어 있지 않습니다. README 설치 가이드를 참조하세요." && exit 1)
	@which protoc-gen-go > /dev/null || go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@which protoc-gen-go-grpc > /dev/null || go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "protoc 환경 확인 완료"

# Proto 파일 다운로드 및 생성
proto: check-protoc
	@echo "Protocol Buffer 파일을 다운로드하고 생성합니다..."
	mkdir -p $(PROTO_OUT_DIR)
	cd cmd/proto-generator && go mod tidy && go run main.go

# Docker 이미지 빌드
docker-build:
	docker build -t $(BINARY_NAME):latest .

# Docker Compose 실행
docker-compose-up:
	docker-compose up -d --build --remove-orphans --force-recreate
	@echo "Docker Compose가 실행되었습니다."
up-db:
	docker-compose up -d --build --remove-orphans --force-recreate pg
	@echo "Docker Compose DB가 실행되었습니다."
up-be:
	docker-compose up -d --build --remove-orphans --force-recreate telemetry-backend
	@echo "Docker Compose BE가 실행되었습니다."


# Docker Compose 중지
docker-compose-down:
	docker-compose down

# 개발 환경 설정 (의존성 설치 등)
dev-setup:
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "필수 도구 설치 완료"
	@echo "주의: protoc 컴파일러는 시스템에 맞게 별도로 설치해야 합니다:"
	@echo "  - Ubuntu/Debian: apt-get install -y protobuf-compiler"
	@echo "  - macOS: brew install protobuf"
	@echo "  - Windows: GitHub에서 릴리스 파일 다운로드"

# 실행 로그 확인
logs:
	docker-compose logs -f

# DB 마이그레이션 실행 (필요시)
db-init:
	$(GO_BUILD_ENV) go run ./cmd/app/main.go --init-db-only

# 전체 초기화 및 실행
all: clean proto build docker-compose-up
	@echo "Application is now running!"