package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/seongpil0948/otel-kafka-pg/modules/common/config"
	"github.com/seongpil0948/otel-kafka-pg/modules/common/logger"
	"github.com/seongpil0948/otel-kafka-pg/modules/common/redis"
)

var (
	ErrCacheMiss = errors.New("cache miss")
)

// CacheService는 캐시 서비스 인터페이스입니다.
type CacheService interface {
	// Get은 캐시에서 데이터를 가져옵니다.
	Get(ctx context.Context, key string, dest interface{}) error

	// Set은 데이터를 캐시에 저장합니다.
	Set(ctx context.Context, key string, value interface{}) error

	// Delete는 캐시에서 데이터를 삭제합니다.
	Delete(ctx context.Context, key string) error

	// IsEnabled는 캐싱이 활성화되어 있는지 여부를 반환합니다.
	IsEnabled() bool
}

// RedisCacheService는 Redis를 사용한 캐시 서비스 구현체입니다.
type RedisCacheService struct {
	redisClient redis.Client
	ttl         time.Duration
	isEnabled   bool
	log         logger.Logger
}

// NewCacheService는 새 캐시 서비스 인스턴스를 생성합니다.
func NewCacheService() (CacheService, error) {
	cfg := config.GetConfig()
	log := logger.GetLogger()

	// Redis 캐싱이 비활성화된 경우 더미 서비스 반환
	if !cfg.Redis.EnableCache {
		log.Info().Msg("Redis 캐싱이 비활성화되어 있습니다")
		return &DummyCacheService{}, nil
	}

	// Redis 클라이언트 생성
	redisClient, err := redis.GetInstance()
	if err != nil {
		log.Error().Err(err).Msg("Redis 클라이언트 생성 실패")
		return nil, err
	}

	return &RedisCacheService{
		redisClient: redisClient,
		ttl:         time.Duration(cfg.Redis.TTL) * time.Second,
		isEnabled:   true,
		log:         log,
	}, nil
}

// Get은 캐시에서 데이터를 가져와 dest에 언마샬링합니다.
func (r *RedisCacheService) Get(ctx context.Context, key string, dest interface{}) error {
	// 캐싱이 비활성화된 경우
	if !r.isEnabled {
		return ErrCacheMiss
	}

	// Redis에서 데이터 가져오기
	data, err := r.redisClient.Get(ctx, key)
	if err != nil {
		r.log.Debug().Err(err).Str("key", key).Msg("캐시 미스")
		return ErrCacheMiss
	}

	// JSON 디코딩
	if err := json.Unmarshal([]byte(data), dest); err != nil {
		r.log.Error().Err(err).Str("key", key).Msg("캐시 데이터 언마샬링 실패")
		return err
	}

	r.log.Debug().Str("key", key).Msg("캐시 히트")
	return nil
}

// Set은 value를 마샬링하여 캐시에 저장합니다.
func (r *RedisCacheService) Set(ctx context.Context, key string, value interface{}) error {
	// 캐싱이 비활성화된 경우
	if !r.isEnabled {
		return nil
	}

	// JSON 인코딩
	data, err := json.Marshal(value)
	if err != nil {
		r.log.Error().Err(err).Str("key", key).Msg("캐시 데이터 마샬링 실패")
		return err
	}

	// Redis에 데이터 저장
	if err := r.redisClient.Set(ctx, key, data, r.ttl); err != nil {
		r.log.Error().Err(err).Str("key", key).Msg("캐시 저장 실패")
		return err
	}

	r.log.Debug().Str("key", key).Int("ttl_seconds", int(r.ttl.Seconds())).Msg("캐시 저장 성공")
	return nil
}

// Delete는 캐시에서 키를 삭제합니다.
func (r *RedisCacheService) Delete(ctx context.Context, key string) error {
	// 캐싱이 비활성화된 경우
	if !r.isEnabled {
		return nil
	}

	// Redis에서 키 삭제
	if err := r.redisClient.Delete(ctx, key); err != nil {
		r.log.Error().Err(err).Str("key", key).Msg("캐시 삭제 실패")
		return err
	}

	r.log.Debug().Str("key", key).Msg("캐시 삭제 성공")
	return nil
}

// IsEnabled는 캐싱이 활성화되어 있는지 여부를 반환합니다.
func (r *RedisCacheService) IsEnabled() bool {
	return r.isEnabled
}

// DummyCacheService는 캐싱을 수행하지 않는 더미 서비스입니다.
type DummyCacheService struct{}

// Get은 항상 캐시 미스를 반환합니다.
func (d *DummyCacheService) Get(ctx context.Context, key string, dest interface{}) error {
	return ErrCacheMiss
}

// Set은 아무 작업도 수행하지 않습니다.
func (d *DummyCacheService) Set(ctx context.Context, key string, value interface{}) error {
	return nil
}

// Delete는 아무 작업도 수행하지 않습니다.
func (d *DummyCacheService) Delete(ctx context.Context, key string) error {
	return nil
}

// IsEnabled는 항상 false를 반환합니다.
func (d *DummyCacheService) IsEnabled() bool {
	return false
}

// GenerateKey는 캐시 키를 생성합니다.
func GenerateKey(prefix string, params ...interface{}) string {
	return fmt.Sprintf("%s:%v", prefix, params)
}
