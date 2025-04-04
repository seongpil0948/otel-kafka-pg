# OpenTelemetry Kafka to PostgreSQL Bridge

이 프로젝트는 OpenTelemetry 텔레메트리 데이터(로그, 트레이스, 메트릭)를 Kafka에서 수신하여 PostgreSQL 데이터베이스에 저장하는 백엔드 서비스입니다.

## 주요 기능

- Kafka에서 OpenTelemetry 프로토콜(OTLP) 포맷의 텔레메트리 데이터 수신
- 수신된 데이터를 파싱하고 처리
- PostgreSQL 데이터베이스에 효율적으로 저장
- 로그와 트레이스 데이터의 배치 처리 및 버퍼링 지원
- 헬스체크 엔드포인트 제공

## 시스템 요구사항

- Go 1.24 이상
- PostgreSQL 13 이상
- Docker 및 Docker Compose (선택사항)
- Kafka (선택사항 - Docker Compose에 포함 가능)

## 설치 및 실행 방법

### 사전 요구사항

1. **Go 설치**: Go 1.24 이상 버전이 필요합니다. [Go 다운로드 페이지](https://golang.org/dl/)에서 설치할 수 있습니다.

2. **Protocol Buffers 컴파일러(protoc) 설치**:
   - **Ubuntu/Debian**:
     ```bash
     apt-get install -y protobuf-compiler
     ```
   - **macOS**:
     ```bash
     brew install protobuf
     ```
   - **Windows**: [GitHub 릴리스 페이지](https://github.com/protocolbuffers/protobuf/releases)에서 다운로드

3. **Go protoc 플러그인 설치**:
   ```bash
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
   ```

### 로컬 환경에서 실행

1. 저장소 클론
   ```bash
   git clone https://github.com/username/otel-kafka-pg.git
   cd otel-kafka-pg
   ```

2. 개발 환경 설정
   ```bash
   make dev-setup
   ```

3. 프로토콜 버퍼 파일 생성
   ```bash
   make proto
   ```

4. 애플리케이션 빌드
   ```bash
   make build
   ```

5. 설정 파일 수정
   `.env` 파일을 생성하거나 수정하여 환경 변수 설정

6. 애플리케이션 실행
   ```bash
   make run
   ```

### Docker Compose로 실행

1. 환경 변수 설정
   `.env` 파일을 프로젝트 루트에 생성하고 필요한 환경 변수 설정

2. Docker Compose로 실행
   ```bash
   make docker-compose-up
   ```

3. 로그 확인
   ```bash
   make logs
   ```

4. 서비스 중지
   ```bash
   make docker-compose-down
   ```

## 환경 변수 설정

`.env` 파일 예시:

```
# 로깅 설정
LOG_LEVEL=info
NODE_ENV=production

# PostgreSQL 연결 설정
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=telemetry
POSTGRES_MAX_CONNECTIONS=20

# Kafka 설정
KAFKA_BROKERS=localhost:9092
KAFKA_GROUP_ID=telemetry-processor-group
KAFKA_TRACE_TOPIC=otlp.traces
KAFKA_LOG_TOPIC=otlp.logs
KAFKA_CLIENT_ID=telemetry-processor

# 배치 처리 설정
BATCH_SIZE=100
FLUSH_INTERVAL=5000
```

## 프로젝트 구조

```
.
├── cmd/                    # 실행 가능한 애플리케이션
│   ├── app/                # 메인 애플리케이션
│   └── healthcheck/        # 헬스체크 유틸리티
├── modules/                # 모듈식 코드 구조
│   ├── common/             # 공통 유틸리티 모듈
│   │   ├── config/         # 설정 관리
│   │   ├── db/             # 데이터베이스 공통 코드
│   │   └── logger/         # 로깅 서비스
│   ├── kafka/              # Kafka 관련 모듈
│   │   ├── consumer/       # Kafka 소비자
│   │   └── processor/      # 메시지 처리 로직
│   ├── log/                # 로그 처리 모듈
│   │   ├── domain/         # 로그 도메인 모델
│   │   ├── repository/     # 로그 저장소
│   │   └── service/        # 로그 서비스
│   └── trace/              # 트레이스 처리 모듈
│       ├── domain/         # 트레이스 도메인 모델
│       ├── repository/     # 트레이스 저장소
│       └── service/        # 트레이스 서비스
├── proto/                  # 생성된 프로토콜 버퍼 파일
├── scripts/                # 유틸리티 스크립트
├── docker-compose.yml      # Docker Compose 설정
├── Dockerfile              # Docker 이미지 빌드 파일
├── go.mod                  # Go 모듈 정의
├── go.work                 # Go 워크스페이스 정의
├── Makefile                # 빌드 및 개발 태스크
└── README.md               # 프로젝트 문서
```

## 개발 가이드

### 모듈 구조

이 프로젝트는 도메인 주도 설계(DDD) 패턴을 따르는 모듈식 구조로 구성되어 있습니다:

- **Domain**: 비즈니스 모델과 규칙 정의
- **Repository**: 데이터 액세스 로직 (PostgreSQL)
- **Service**: 비즈니스 로직 구현
- **Processor**: 메시지 처리 및 변환

### 새 기능 추가하기

1. 필요한 도메인 모델 생성 또는 수정
2. 레포지토리 계층에 데이터 액세스 메서드 구현
3. 서비스 계층에 비즈니스 로직 구현
4. 필요시 메시지 처리기 추가 또는 수정

### 주요 컴포넌트

1. **Kafka Consumer**: Kafka에서 메시지를 수신하고 적절한 처리기로 라우팅
2. **Message Processor**: 수신된 메시지를 파싱하고 변환
3. **Buffering System**: 데이터베이스 효율성을 위한 메시지 그룹화
4. **Repositories**: PostgreSQL에 데이터 저장
5. **Health Check**: 시스템 상태 모니터링

#### HI
```bash
# kafka 모듈의 go.mod 파일 확인
cd modules/kafka
# go.mod 파일이 없다면 생성
go mod init github.com/seongpil0948/otel-kafka-pg/modules/kafka
# 의존성 정리
go mod tidy
```