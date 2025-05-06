FROM golang:1.24-bookworm AS builder

RUN apt-get update && apt-get install -y git build-essential

# 작업 디렉토리 설정
WORKDIR /app

# Go workspace 설정
COPY . .

# 모듈 초기화 및 의존성 설치
RUN ./scripts/tidy-go-modules.sh

# 빌드
WORKDIR /app
RUN CGO_ENABLED=1 go build -o /go/bin/app ./cmd/app/main.go
RUN CGO_ENABLED=1 go build -o /go/bin/healthcheck ./cmd/healthcheck/main.go

# 최종 이미지
FROM debian:stable-slim

# 필수 패키지 설치
RUN apt-get update && apt-get install -y ca-certificates tzdata redis-tools && rm -rf /var/lib/apt/lists/*

# 사용자 추가
RUN addgroup --system --gid 1001 telemetry && \
  adduser --system --uid 1001 --ingroup telemetry telemetry

WORKDIR /app

# 스크립트 디렉토리 생성
RUN mkdir -p /app/scripts

# 바이너리 및 스크립트 복사
COPY --from=builder /go/bin/app /app/
COPY --from=builder /go/bin/healthcheck /app/
COPY --from=builder /app/scripts /app/scripts

# 스크립트에 실행 권한 부여
RUN chmod +x /app/scripts/*.sh

# 실행 사용자 변경
USER telemetry

# 헬스 체크 설정
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 CMD [ "/app/healthcheck" ]

# 명령어 설정
CMD ["/app/app"]