package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

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
	_ = config.LoadConfig() // 설정 로드만 하고 변수는 직접 참조하지 않음

	// 2. 로거 초기화
	log := logger.Init()
	log.Info().Msg("텔레메트리 백엔드 시작 중...")

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

	// 6. Kafka 프로세서 및 컨슈머 설정
	proc := processor.NewProcessor()
	kafkaConsumer := consumer.NewConsumer(proc, traceSvc, logSvc)

	// 7. Kafka 컨슈머 시작
	log.Info().Msg("Kafka 컨슈머 시작 중...")
	if err := kafkaConsumer.Start(ctx); err != nil {
		log.Fatal().Err(err).Msg("Kafka 컨슈머 시작 실패")
	}
	log.Info().Msg("Kafka 컨슈머가 실행 중입니다")

	// 8. 애플리케이션 상태 로깅
	log.Info().Msg("텔레메트리 백엔드가 정상적으로 실행 중입니다")

	// 종료 시그널 처리
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	// 종료 시그널 대기
	sig := <-sigCh
	log.Info().Str("signal", sig.String()).Msg("종료 신호 수신, 정상 종료를 시작합니다")

	// 9. 정상 종료 처리
	shutdown(ctx, database, kafkaConsumer, log)
}

// shutdown은 애플리케이션을 정상적으로 종료합니다.
func shutdown(ctx context.Context, database commonDB.Database, kafkaConsumer consumer.Consumer, log logger.Logger) {
	// Kafka 컨슈머 종료
	if err := kafkaConsumer.Stop(); err != nil {
		log.Error().Err(err).Msg("Kafka 컨슈머 종료 실패")
	}

	// 데이터베이스 연결 종료
	if err := database.Close(); err != nil {
		log.Error().Err(err).Msg("데이터베이스 연결 종료 실패")
	}

	log.Info().Msg("애플리케이션이 정상적으로 종료되었습니다")
}