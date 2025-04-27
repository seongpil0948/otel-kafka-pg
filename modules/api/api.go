package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/seongpil0948/otel-kafka-pg/modules/api/router"
	"github.com/seongpil0948/otel-kafka-pg/modules/common/config"
	"github.com/seongpil0948/otel-kafka-pg/modules/common/db"
	"github.com/seongpil0948/otel-kafka-pg/modules/common/logger"
	"github.com/seongpil0948/otel-kafka-pg/modules/log/repository"
	logService "github.com/seongpil0948/otel-kafka-pg/modules/log/service"
	traceRepository "github.com/seongpil0948/otel-kafka-pg/modules/trace/repository"
	traceService "github.com/seongpil0948/otel-kafka-pg/modules/trace/service"

	_ "github.com/seongpil0948/otel-kafka-pg/docs"
)

// Server API 서버 구조체
type Server struct {
	Router       *gin.Engine
	HttpServer   *http.Server
	Config       *config.Config
	Log          logger.Logger
	TraceService traceService.TraceService
	LogService   logService.LogService
}

// NewServer는 새 API 서버 인스턴스를 생성합니다
func NewServer(cfg *config.Config, log logger.Logger, database db.Database) *Server {
	// 저장소 생성
	logRepo := repository.NewLogRepository(database)
	traceRepo := traceRepository.NewTraceRepository(database)

	// 서비스 생성
	logSvc := logService.NewLogService(logRepo)
	traceSvc := traceService.NewTraceService(traceRepo)

	// 라우터 설정
	ginRouter := router.SetupRouter(cfg, log, traceSvc, logSvc)

	// HTTP 서버 설정
	httpServer := &http.Server{
		Addr:         ":" + "8080",
		Handler:      ginRouter,
		ReadTimeout:  time.Duration(cfg.API.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.API.WriteTimeout) * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		Router:       ginRouter,
		HttpServer:   httpServer,
		Config:       cfg,
		Log:          log,
		TraceService: traceSvc,
		LogService:   logSvc,
	}
}

// Start는 API 서버를 시작합니다
func (s *Server) Start() error {
	// 서버 시작
	s.Log.Info().Int("port", s.Config.API.Port).Msg("API 서버 시작")

	if err := s.HttpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.Log.Error().Err(err).Msg("API 서버 시작 중 오류 발생")
		return err
	}

	return nil
}

// Stop은 API 서버를 정상적으로 종료합니다
func (s *Server) Stop(ctx context.Context) error {
	s.Log.Info().Msg("API 서버 정상 종료 중...")

	// 서버 종료
	if err := s.HttpServer.Shutdown(ctx); err != nil {
		s.Log.Error().Err(err).Msg("API 서버 정상 종료 실패")
		return err
	}

	s.Log.Info().Msg("API 서버가 정상적으로 종료되었습니다")
	return nil
}
