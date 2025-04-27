package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/seongpil0948/otel-kafka-pg/modules/common/logger"
)

// RequestLogger 미들웨어는 요청 로깅을 수행합니다
func RequestLogger(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// 요청 ID 생성 또는 사용
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			c.Header("X-Request-ID", requestID)
		}

		// 요청 처리 전 로깅
		log.Info().
			Str("request_id", requestID).
			Str("method", method).
			Str("path", path).
			Str("client_ip", c.ClientIP()).
			Msg("API 요청 시작")

		// 요청 처리
		c.Next()

		// 요청 처리 후 로깅
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		logEvent := log.Info()
		if statusCode >= 400 {
			logEvent = log.Error()
		}

		logEvent.
			Str("request_id", requestID).
			Str("method", method).
			Str("path", path).
			Int("status", statusCode).
			Dur("latency", latency).
			Int("bytes", c.Writer.Size()).
			Msg("API 요청 완료")
	}
}

// ErrorHandler 미들웨어는 패닉 발생 시 처리합니다
func ErrorHandler(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 요청 ID 가져오기
				requestID := c.GetHeader("X-Request-ID")

				// 패닉 로깅
				log.Error().
					Str("request_id", requestID).
					Interface("error", err).
					Str("method", c.Request.Method).
					Str("path", c.Request.URL.Path).
					Msg("API 처리 중 패닉 발생")

				// 에러 응답 전송
				c.JSON(500, gin.H{
					"success": false,
					"error": gin.H{
						"code":    500,
						"message": "Internal Server Error",
					},
				})

				c.Abort()
			}
		}()
		c.Next()
	}
}

// CORS 미들웨어 설정
func CORS(allowedOrigins []string, allowCredentials bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 허용된 출처인지 확인
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Request-ID")

			if allowCredentials {
				c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			}
		}

		// OPTIONS 요청 처리
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
