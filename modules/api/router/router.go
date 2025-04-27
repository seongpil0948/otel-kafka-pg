package router

import (
	"fmt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/seongpil0948/otel-kafka-pg/modules/api/controller"
	"github.com/seongpil0948/otel-kafka-pg/modules/api/middleware"
	"github.com/seongpil0948/otel-kafka-pg/modules/common/config"
	"github.com/seongpil0948/otel-kafka-pg/modules/common/logger"
	logService "github.com/seongpil0948/otel-kafka-pg/modules/log/service"
	traceService "github.com/seongpil0948/otel-kafka-pg/modules/trace/service"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter는 API 라우터 및 미들웨어를 설정합니다
func SetupRouter(cfg *config.Config, log logger.Logger, traceService traceService.TraceService, logService logService.LogService) *gin.Engine {
	// 환경에 따른 Gin 모드 설정
	if cfg.Logger.IsDev {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 라우터 생성
	router := gin.New()

	// 미들웨어 설정
	router.Use(middleware.ErrorHandler(log))
	router.Use(middleware.RequestLogger(log))

	// CORS 미들웨어 설정
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = cfg.API.AllowedOrigins
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"}
	corsConfig.AllowCredentials = cfg.API.AllowCredentials
	router.Use(cors.New(corsConfig))

	// 컨트롤러 생성
	traceController := controller.NewTraceController(traceService, log)
	logController := controller.NewLogController(logService, log)

	// 기본 경로 설정
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "OpenTelemetry API 서버가 실행 중입니다",
		})
	})

	// 헬스 체크 엔드포인트
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"version": "1.0.0",
		})
	})

	// API 그룹
	api := router.Group(cfg.API.BasePath)
	{
		// 텔레메트리 API 그룹
		telemetry := api.Group("/telemetry")
		{
			// 트레이스 관련 엔드포인트
			traces := telemetry.Group("/traces")
			{
				traces.GET("", traceController.QueryTraces)
				traces.GET("/:traceId", traceController.GetTraceByID)
			}

			// 로그 관련 엔드포인트
			logs := telemetry.Group("/logs")
			{
				logs.GET("", logController.QueryLogs)
				logs.GET("/trace/:traceId", logController.GetLogsByTraceID)
				logs.GET("/summary", logController.GetLogSummary)
			}

			// 메트릭 관련 엔드포인트
			metrics := telemetry.Group("/metrics")
			{
				metrics.GET("/services", traceController.GetServiceMetrics)
			}
		}
	}

	// Swagger 문서화 설정
	if cfg.API.EnableSwagger {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	return router
}

// StartServer는 API 서버를 시작합니다
func StartServer(cfg *config.Config, router *gin.Engine, log logger.Logger) error {
	addr := fmt.Sprintf("%s:%d", cfg.API.Host, cfg.API.Port)
	log.Info().Int("port", cfg.API.Port).Str("host", cfg.API.Host).Msg("API 서버 시작")
	return router.Run(addr)
}
