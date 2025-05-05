package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/seongpil0948/otel-kafka-pg/modules/common/cache"
	"github.com/seongpil0948/otel-kafka-pg/modules/common/logger"
)

// responseBodyWriter는 응답 본문을 캡처하는 구조체입니다.
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write는 ResponseWriter의 Write 메서드를 오버라이드합니다.
func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// WriteString은 ResponseWriter의 WriteString 메서드를 오버라이드합니다.
func (r responseBodyWriter) WriteString(s string) (int, error) {
	r.body.WriteString(s)
	return r.ResponseWriter.WriteString(s)
}

// CachingMiddleware는 API 응답을 캐싱하는 미들웨어입니다.
// CachingMiddleware는 API 응답을 캐싱하는 미들웨어입니다.
func CachingMiddleware(cacheService cache.CacheService, log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// GET 메서드만 캐싱
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		// 캐싱이 비활성화된 경우
		if !cacheService.IsEnabled() {
			log.Debug().Str("url", c.Request.URL.String()).Msg("캐싱 비활성화 상태")
			c.Next()
			return
		}

		// 요청 본문 유지를 위해 복사
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = ioutil.ReadAll(c.Request.Body)
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 캐시 키 생성
		cacheKey := generateCacheKey(c.Request, requestBody)
		log.Debug().Str("cache_key", cacheKey).Str("url", c.Request.URL.String()).Msg("캐시 키 생성")

		// 캐시에서 응답 확인
		var cachedResponse struct {
			Status int                 `json:"status"`
			Header map[string][]string `json:"header"`
			Data   []byte              `json:"data"`
		}

		err := cacheService.Get(c.Request.Context(), cacheKey, &cachedResponse)
		if err == nil {
			// 캐시 히트: 캐시된 응답 반환
			log.Info().Str("url", c.Request.URL.String()).Str("cache_key", cacheKey).Msg("캐시 히트")
			c.Header("X-Cache", "HIT")

			// 캐시된 헤더 설정
			for key, values := range cachedResponse.Header {
				for _, value := range values {
					c.Writer.Header().Add(key, value)
				}
			}

			// 캐시된 응답 본문 작성
			c.Writer.WriteHeader(cachedResponse.Status)
			c.Writer.Write(cachedResponse.Data)
			c.Abort()
			return
		}

		// 캐시 미스: 응답 캡처
		log.Info().Str("url", c.Request.URL.String()).Str("cache_key", cacheKey).Err(err).Msg("캐시 미스")
		c.Header("X-Cache", "MISS")
		responseBody := &bytes.Buffer{}
		rbw := &responseBodyWriter{ResponseWriter: c.Writer, body: responseBody}
		c.Writer = rbw

		c.Next()

		// 응답이 이미 전송되었고 상태 코드가 성공인 경우에만 캐싱
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			// 캐시에 응답 저장
			response := struct {
				Status int                 `json:"status"`
				Header map[string][]string `json:"header"`
				Data   []byte              `json:"data"`
			}{
				Status: c.Writer.Status(),
				Header: c.Writer.Header().Clone(),
				Data:   responseBody.Bytes(),
			}

			if err := cacheService.Set(c.Request.Context(), cacheKey, response); err != nil {
				log.Error().Err(err).Str("url", c.Request.URL.String()).Str("cache_key", cacheKey).Msg("응답 캐싱 실패")
			} else {
				log.Info().Str("url", c.Request.URL.String()).Str("cache_key", cacheKey).Msg("응답 캐싱 성공")
			}
		} else {
			log.Debug().Str("url", c.Request.URL.String()).Int("status", c.Writer.Status()).Msg("응답 캐싱 건너뜀: 성공 상태 코드가 아님")
		}
	}
}

// generateCacheKey는 요청에 대한 고유한 캐시 키를 생성합니다.
func generateCacheKey(req *http.Request, body []byte) string {
	// URI 포함
	path := req.URL.Path

	// 쿼리 파라미터 정렬 포함
	query := req.URL.Query()

	// 정렬된 쿼리 파라미터 문자열 생성
	var queryParts []string
	if len(query) > 0 {
		for k, values := range query {
			if len(values) == 1 {
				queryParts = append(queryParts, fmt.Sprintf("%s=%s", k, values[0]))
			} else {
				// 복수 값이 있는 경우 정렬하여 추가
				sort.Strings(values)
				for _, v := range values {
					queryParts = append(queryParts, fmt.Sprintf("%s=%s", k, v))
				}
			}
		}
		sort.Strings(queryParts)
	}

	// 해시 생성
	h := sha256.New()
	h.Write([]byte(path))

	// 쿼리 파라미터 해시에 추가
	for _, part := range queryParts {
		h.Write([]byte(part))
	}

	// 요청 본문이 있는 경우 해시 추가
	if len(body) > 0 {
		h.Write(body)
	}

	// 사용자 에이전트 정보 (브라우저 캐시 구분용)
	userAgent := req.Header.Get("User-Agent")
	if userAgent != "" {
		h.Write([]byte(userAgent))
	}

	// 키 접두사 + 해시
	hashStr := hex.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("api:cache:%s:%s", path, hashStr[:16])
}

// InvalidateCacheMiddleware는 변경 작업(POST, PUT, DELETE 등) 후 관련 캐시를 무효화하는 미들웨어입니다.
func InvalidateCacheMiddleware(cacheService cache.CacheService, log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// GET 요청은 무시
		if c.Request.Method == http.MethodGet {
			c.Next()
			return
		}

		// 캐싱이 비활성화된 경우
		if !cacheService.IsEnabled() {
			c.Next()
			return
		}

		// 핸들러 실행
		c.Next()

		// 요청이 성공한 경우에만 캐시 무효화
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			// 현재는 특정 패턴의 캐시만 무효화함
			// 실제 구현에서는 더 정교한 방식으로 관련 캐시를 식별하여 무효화해야 함

			// 무효화할 패턴 추출 (예: /api/v1/logs -> api:cache:/api/v1/logs)
			invalidationPattern := "api:cache:" + c.Request.URL.Path

			// 알림 로깅
			log.Info().Str("pattern", invalidationPattern).Msg("캐시 무효화")

			// 여기서는 패턴과 정확히 일치하는 캐시만 무효화 (실제로는 패턴 기반 무효화 구현 필요)
			if err := cacheService.Delete(c.Request.Context(), invalidationPattern); err != nil {
				log.Error().Err(err).Str("pattern", invalidationPattern).Msg("캐시 무효화 실패")
			}

			// 패턴이 '/api/v1/logs/123'이면 '/api/v1/logs' 패턴도 무효화
			if strings.Count(c.Request.URL.Path, "/") > 2 {
				parts := strings.Split(c.Request.URL.Path, "/")
				parentPattern := "api:cache:" + strings.Join(parts[:len(parts)-1], "/")
				if err := cacheService.Delete(c.Request.Context(), parentPattern); err != nil {
					log.Error().Err(err).Str("pattern", parentPattern).Msg("부모 패턴 캐시 무효화 실패")
				}
			}
		}
	}
}
