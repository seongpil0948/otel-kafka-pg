package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/seongpil0948/otel-kafka-pg/modules/cleanup" // 추가된 import
	"github.com/seongpil0948/otel-kafka-pg/modules/common/config"
	commonDB "github.com/seongpil0948/otel-kafka-pg/modules/common/db"
	"github.com/seongpil0948/otel-kafka-pg/modules/common/logger"
	"github.com/seongpil0948/otel-kafka-pg/modules/kafka/consumer"
	"github.com/seongpil0948/otel-kafka-pg/modules/kafka/processor"
	"github.com/seongpil0948/otel-kafka-pg/modules/log/repository"
	logService "github.com/seongpil0948/otel-kafka-pg/modules/log/service"
	traceRepository "github.com/seongpil0948/otel-kafka-pg/modules/trace/repository"
	traceService "github.com/seongpil0948/otel-kafka-pg/modules/trace/service"
)

func main() {
	// 컨텍스트 생성
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. 설정 초기화
	cfg := config.LoadConfig() // 설정 로드 및 변수 참조

	// 2. 로거 초기화
	log := logger.Init()
	log.Info().Msg("OpenTelemetry 텔레메트리 백엔드 시작 중...")

	// 3. 데이터베이스 연결 설정
	log.Info().Msg("데이터베이스 연결 초기화 중...")
	database, err := commonDB.NewDatabase()
	if err != nil {
		log.Fatal().Err(err).Msg("데이터베이스 연결 실패")
	}
	log.Info().Msg("데이터베이스 연결 성공")

	// 4. 데이터베이스 스키마 확인 및 초기화
	initialized, err := commonDB.IsDatabaseInitialized(database)
	if err != nil {
		log.Fatal().Err(err).Msg("데이터베이스 초기화 확인 실패")
	}

	if !initialized {
		log.Info().Msg("데이터베이스 스키마가 존재하지 않음, 초기화 시작...")
		if err := commonDB.InitializeSchema(database); err != nil {
			log.Fatal().Err(err).Msg("데이터베이스 스키마 초기화 실패")
		}
		log.Info().Msg("데이터베이스 스키마 초기화 완료")
	} else {
		log.Info().Msg("데이터베이스 스키마가 이미 초기화되어 있음")
	}

	// 5. 저장소 및 서비스 계층 설정
	logRepo := repository.NewLogRepository(database)
	traceRepo := traceRepository.NewTraceRepository(database)
	
	logSvc := logService.NewLogService(logRepo)
	traceSvc := traceService.NewTraceService(traceRepo)

	// 6. 데이터 정리 서비스 설정 및 시작
	cleanupSvc := cleanup.NewCleanupService(database, cfg)
	if err := cleanupSvc.Start(ctx); err != nil {
		log.Error().Err(err).Msg("데이터 정리 서비스 시작 실패")
	}

	// 7. Kafka 프로세서 및 컨슈머 설정
	proc := processor.NewProcessor()
	kafkaConsumer := consumer.NewConsumer(proc, traceSvc, logSvc)

	// 8. Kafka 컨슈머 시작
	log.Info().Msg("Kafka 컨슈머 시작 중...")
	if err := kafkaConsumer.Start(ctx); err != nil {
		log.Fatal().Err(err).Msg("Kafka 컨슈머 시작 실패")
	}
	log.Info().Msg("Kafka 컨슈머가 실행 중입니다")

	// 9. 애플리케이션 상태 로깅
	log.Info().Msg("OpenTelemetry 텔레메트리 백엔드가 정상적으로 실행 중입니다")
	log.Info().Msg("프로토콜 버퍼를 사용하여 로그 및 트레이스 데이터 처리 중...")

	// 종료 시그널 처리
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	// 종료 시그널 대기
	sig := <-sigCh
	log.Info().Str("signal", sig.String()).Msg("종료 신호 수신, 정상 종료를 시작합니다")

	// 10. 정상 종료 처리 
	shutdown(ctx, database, kafkaConsumer, cleanupSvc, log)
}

// shutdown은 애플리케이션을 정상적으로 종료합니다.
func shutdown(ctx context.Context, database commonDB.Database, kafkaConsumer consumer.Consumer, cleanupSvc cleanup.CleanupService, log logger.Logger) {
	// 데이터 정리 서비스 종료
	log.Info().Msg("데이터 정리 서비스 종료 중...")
	if err := cleanupSvc.Stop(); err != nil {
		log.Error().Err(err).Msg("데이터 정리 서비스 종료 실패")
	} else {
		log.Info().Msg("데이터 정리 서비스가 정상적으로 종료되었습니다")
	}

	// Kafka 컨슈머 종료
	log.Info().Msg("Kafka 컨슈머 종료 중...")
	if err := kafkaConsumer.Stop(); err != nil {
		log.Error().Err(err).Msg("Kafka 컨슈머 종료 실패")
	} else {
		log.Info().Msg("Kafka 컨슈머가 정상적으로 종료되었습니다")
	}

	// 데이터베이스 연결 종료
	log.Info().Msg("데이터베이스 연결 종료 중...")
	if err := database.Close(); err != nil {
		log.Error().Err(err).Msg("데이터베이스 연결 종료 실패")
	} else {
		log.Info().Msg("데이터베이스 연결이 정상적으로 종료되었습니다")
	}

	log.Info().Msg("애플리케이션이 정상적으로 종료되었습니다")
}