package redis

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/seongpil0948/otel-kafka-pg/modules/common/config"
	"github.com/seongpil0948/otel-kafka-pg/modules/common/logger"
)

// Client는 Redis 클라이언트 인터페이스입니다.
type Client interface {
	// Get는 키에 해당하는 값을 가져옵니다.
	Get(ctx context.Context, key string) (string, error)

	// Set은 키-값을 설정하고 만료 시간을 지정합니다.
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error

	// Delete는 키를 삭제합니다.
	Delete(ctx context.Context, key string) error

	// Close는 Redis 연결을 종료합니다.
	Close() error
}

// RedisClient는 Redis 클라이언트 구현체입니다.
type RedisClient struct {
	client *redis.Client
	log    logger.Logger
}

var (
	instance *RedisClient
	once     sync.Once
)

// NewRedisClient는 새 Redis 클라이언트 인스턴스를 생성합니다.
// NewRedisClient는 새 Redis 클라이언트 인스턴스를 생성합니다.
func NewRedisClient() (Client, error) {
	var err error
	var retryCount = 0
	var maxRetries = 3

	once.Do(func() {
		cfg := config.GetConfig()
		log := logger.GetLogger()

		for retryCount < maxRetries {
			client := redis.NewClient(&redis.Options{
				Addr:     cfg.Redis.Address,
				Password: cfg.Redis.Password,
				DB:       cfg.Redis.DB,
				PoolSize: cfg.Redis.PoolSize,
			})

			// 연결 테스트
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err = client.Ping(ctx).Err(); err != nil {
				retryCount++
				log.Warn().Err(err).Int("retry", retryCount).Msg("Redis 연결 실패, 재시도 중...")
				time.Sleep(time.Duration(retryCount) * time.Second) // 지수 백오프
				continue
			}

			instance = &RedisClient{
				client: client,
				log:    log,
			}

			log.Info().
				Str("addr", cfg.Redis.Address).
				Int("db", cfg.Redis.DB).
				Int("poolSize", cfg.Redis.PoolSize).
				Msg("Redis 연결 성공")

			return
		}

		log.Error().Err(err).Int("retries", retryCount).Msg("최대 재시도 횟수에 도달했습니다. Redis 연결 실패")
	})

	if err != nil {
		return nil, err
	}

	return instance, nil
}

// GetInstance는 싱글톤 인스턴스를 반환합니다.
func GetInstance() (Client, error) {
	if instance == nil {
		return NewRedisClient()
	}
	return instance, nil
}

// Get는 키에 해당하는 값을 가져옵니다.
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

// Set은 키-값을 설정하고 만료 시간을 지정합니다.
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Delete는 키를 삭제합니다.
func (r *RedisClient) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Close는 Redis 연결을 종료합니다.
func (r *RedisClient) Close() error {
	r.log.Info().Msg("Redis 연결 종료 중...")
	err := r.client.Close()
	if err != nil {
		return err
	}
	r.log.Info().Msg("Redis 연결 종료됨")
	return nil
}
