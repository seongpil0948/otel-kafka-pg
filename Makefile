.PHONY: build run test clean proto docker-build docker compose-up docker compose-down tidy redis-cli

-include .env
export

# 기본 변수 설정
BINARY_NAME=telemetry-backend
GO_BUILD_ENV=CGO_ENABLED=1
GO_FILES=$(shell find . -type f -name "*.go" -not -path "./proto/*")

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
	go clean -modcache

# 모든 모듈 go mod tidy 실행
tidy:
	./scripts/tidy-go-modules.sh

# Redis CLI 실행
redis-cli:
	docker exec -it telemetry-redis redis-cli

# Redis 모니터링
redis-monitor:
	docker exec -it telemetry-redis redis-cli monitor

# Docker 이미지 빌드
docker-build:
	docker build -t $(BINARY_NAME):latest .

# Docker Compose 실행
docker-compose-up:
	docker compose up -d --build --remove-orphans --force-recreate
	@echo "Docker Compose가 실행되었습니다."

# Redis만 실행
up-redis:
	docker compose up -d --build --remove-orphans --force-recreate redis
	@echo "Redis가 실행되었습니다."

# DB만 실행
up-db:
	docker compose up -d --build --remove-orphans --force-recreate pg
	@echo "PostgreSQL DB가 실행되었습니다."

# 백엔드만 실행
up-be:
	docker compose up -d --build --remove-orphans --force-recreate telemetry-backend
	@echo "백엔드가 실행되었습니다."

# Docker Compose 중지
docker compose-down:
	docker compose down

# 개발 환경 설정 (의존성 설치 등)
dev-setup:
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	chmod +x ./scripts/*.sh
	@echo "필수 도구 설치 완료"
	@echo "  - Windows: GitHub에서 릴리스 파일 다운로드"

# 실행 로그 확인
logs:
	docker compose logs -f

# 백엔드 로그만 확인
logs-be:
	docker compose logs -f telemetry-backend

# Redis 로그만 확인
logs-redis:
	docker compose logs -f redis

# DB 마이그레이션 실행 (필요시)
db-init:
	$(GO_BUILD_ENV) go run ./cmd/app/main.go --init-db-only

# Redis 캐시 확인
check-cache:
	./scripts/check-cache-env.sh

# Swagger 문서 생성
swagger:
	mkdir -p docs
	swag init -g cmd/app/main.go -o docs --parseDependency --parseInternal
	@echo "Swagger 문서가 생성되었습니다."

# Swagger 문서 서식 정리
swagger-fmt:
	swag fmt -g cmd/app/main.go -d ./
	@echo "Swagger 문서 서식이 정리되었습니다."
	
# 전체 초기화 및 실행
all: clean build docker compose-up
	@echo "Application is now running!"

push-gitlab:
	git push gitlab main:prd 
	@echo "GitLab에 푸시 완료"