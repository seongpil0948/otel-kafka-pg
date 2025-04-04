FROM golang:1.24-alpine AS builder

# 필수 패키지 설치
RUN apk add --no-cache git gcc libc-dev

# 작업 디렉토리 설정
WORKDIR /app

# Go workspace 설정
COPY go.work go.work

# 모듈 복사
COPY modules/ modules/
COPY cmd/ cmd/

# 모듈 초기화
WORKDIR /app/modules/common
RUN go mod download
WORKDIR /app/modules/log
RUN go mod download
WORKDIR /app/modules/trace
RUN go mod download
WORKDIR /app/modules/kafka
RUN go mod download
WORKDIR /app/cmd/app
RUN go mod download
WORKDIR /app/cmd/healthcheck
RUN go mod download

# 빌드
WORKDIR /app
RUN CGO_ENABLED=1 go build -o /go/bin/app ./cmd/app/main.go
RUN CGO_ENABLED=1 go build -o /go/bin/healthcheck ./cmd/healthcheck/main.go

# 최종 이미지
FROM alpine:latest

# 필수 패키지 설치
RUN apk add --no-cache ca-certificates tzdata libc6-compat

# 사용자 추가
RUN addgroup --system --gid 1001 telemetry && \
  adduser --system --uid 1001 --ingroup telemetry telemetry

WORKDIR /app

# 바이너리 복사
COPY --from=builder /go/bin/app /app/
COPY --from=builder /go/bin/healthcheck /app/

# 실행 사용자 변경
USER telemetry

# 헬스 체크 설정
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 CMD [ "/app/healthcheck" ]

# 명령어 설정
CMD ["/app/app"]