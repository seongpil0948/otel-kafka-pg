package cleanup

import (
	"context"
	"fmt"
	"time"

	"github.com/seongpil0948/otel-kafka-pg/modules/common/config"
	"github.com/seongpil0948/otel-kafka-pg/modules/common/db"
	"github.com/seongpil0948/otel-kafka-pg/modules/common/logger"
)

// CleanupService는 오래된 텔레메트리 데이터를 정리하는 서비스입니다.
type CleanupService interface {
	// Start는 데이터 정리 서비스를 시작합니다.
	Start(ctx context.Context) error
	// Stop은 데이터 정리 서비스를 중지합니다.
	Stop() error
}

// cleanupServiceImpl은 CleanupService 인터페이스의 구현체입니다.
type cleanupServiceImpl struct {
	db            db.Database
	config        *config.Config
	log           logger.Logger
	ticker        *time.Ticker
	stopChan      chan struct{}
	isRunning     bool
}

// NewCleanupService는 새 CleanupService 인스턴스를 생성합니다.
func NewCleanupService(database db.Database, config *config.Config) CleanupService {
	return &cleanupServiceImpl{
		db:            database,
		config:        config,
		log:           logger.GetLogger(),
		stopChan:      make(chan struct{}),
		isRunning:     false,
	}
}

// Start는 데이터 정리 서비스를 시작합니다.
func (c *cleanupServiceImpl) Start(ctx context.Context) error {
	if c.isRunning {
		c.log.Info().Msg("데이터 정리 서비스가 이미 실행 중입니다")
		return nil
	}

	if !c.config.DataRetention.Enabled {
		c.log.Info().Msg("데이터 정리 서비스가 비활성화되어 있습니다")
		return nil
	}

	interval := time.Duration(c.config.DataRetention.CleanupInterval) * time.Minute
	c.ticker = time.NewTicker(interval)
	c.isRunning = true

	c.log.Info().
		Int("interval_minutes", c.config.DataRetention.CleanupInterval).
		Int("retention_days", c.config.DataRetention.RetentionPeriod).
		Msg("데이터 정리 서비스 시작")

	// 시작 시 즉시 한 번 실행
	if err := c.cleanupOldData(); err != nil {
		c.log.Error().Err(err).Msg("초기 데이터 정리 중 오류 발생")
	}

	go func() {
		for {
			select {
			case <-c.ticker.C:
				if err := c.cleanupOldData(); err != nil {
					c.log.Error().Err(err).Msg("데이터 정리 중 오류 발생")
				}
			case <-c.stopChan:
				c.log.Info().Msg("데이터 정리 서비스 루프 종료")
				return
			case <-ctx.Done():
				c.log.Info().Msg("컨텍스트 종료로 인한 데이터 정리 서비스 루프 종료")
				return
			}
		}
	}()

	return nil
}

// Stop은 데이터 정리 서비스를 중지합니다.
func (c *cleanupServiceImpl) Stop() error {
	if !c.isRunning {
		return nil
	}

	c.ticker.Stop()
	c.stopChan <- struct{}{}
	c.isRunning = false
	c.log.Info().Msg("데이터 정리 서비스 중지됨")
	return nil
}

func (c *cleanupServiceImpl) cleanupOldData() error {
	retentionDays := c.config.DataRetention.RetentionPeriod
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays).UnixNano() / 1000000 // 밀리초 단위로 변환

	c.log.Info().
		Int("retention_days", retentionDays).
		Int64("cutoff_time_ms", cutoffTime).
		Msg("오래된 데이터 정리 시작")

	startTime := time.Now()

	// 트랜잭션 시작
	tx, err := c.db.Begin()
	if err != nil {
		return fmt.Errorf("데이터 정리 트랜잭션 시작 실패: %w", err)
	}

	// 롤백 함수 준비
	rollbackOnError := func() {
		if err != nil {
			tx.Rollback()
			c.log.Error().Err(err).Msg("데이터 정리 트랜잭션 롤백됨")
		}
	}
	defer rollbackOnError()

	// 로그 삭제
	logResult, err := tx.Exec("DELETE FROM logs WHERE timestamp < $1", cutoffTime)
	if err != nil {
		return fmt.Errorf("로그 정리 실패: %w", err)
	}

	logCount, err := logResult.RowsAffected()
	if err != nil {
		c.log.Warn().Err(err).Msg("삭제된 로그 행 수를 가져올 수 없습니다")
	}

	// 트레이스 삭제
	traceResult, err := tx.Exec("DELETE FROM traces WHERE start_time < $1", cutoffTime)
	if err != nil {
		return fmt.Errorf("트레이스 정리 실패: %w", err)
	}

	traceCount, err := traceResult.RowsAffected()
	if err != nil {
		c.log.Warn().Err(err).Msg("삭제된 트레이스 행 수를 가져올 수 없습니다")
	}

	// 메트릭 삭제 (메트릭 테이블이 있는 경우)
	metricResult, err := tx.Exec("DELETE FROM metrics WHERE timestamp < $1", cutoffTime)
	if err != nil {
		// 메트릭 테이블이 없을 수 있으므로 오류를 무시하고 로그만 남깁니다
		c.log.Debug().Err(err).Msg("메트릭 테이블이 없거나 정리 중 오류 발생")
		metricResult = nil
	}

	var metricCount int64 = 0
	if metricResult != nil {
		metricCount, err = metricResult.RowsAffected()
		if err != nil {
			c.log.Warn().Err(err).Msg("삭제된 메트릭 행 수를 가져올 수 없습니다")
		}
	}

	// 트랜잭션 커밋
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("데이터 정리 트랜잭션 커밋 실패: %w", err)
	}

	duration := time.Since(startTime)
	c.log.Info().
		Int64("logs_deleted", logCount).
		Int64("traces_deleted", traceCount).
		Int64("metrics_deleted", metricCount).
		Dur("duration", duration).
		Msg("데이터 정리 완료")

	return nil
}