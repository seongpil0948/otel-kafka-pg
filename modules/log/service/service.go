package service

import (
	"github.com/seongpil0948/otel-kafka-pg/modules/common/logger"
	"github.com/seongpil0948/otel-kafka-pg/modules/log/domain"
	"github.com/seongpil0948/otel-kafka-pg/modules/log/repository"
)

// LogService는 로그 서비스 인터페이스입니다.
type LogService interface {
	// 로그 저장
	SaveLogs(logs []domain.LogItem) error
	
	// 로그 쿼리
	QueryLogs(filter domain.LogFilter) (domain.LogQueryResult, error)
	
	// 로그 집계
	GetServiceAggregation(startTime, endTime int64) ([]domain.ServiceAggregation, error)
	GetSeverityAggregation(startTime, endTime int64) ([]domain.SeverityAggregation, error)
}

// LogServiceImpl은 로그 서비스 구현체입니다.
type LogServiceImpl struct {
	repository repository.LogRepository
	log        logger.Logger
}

// NewLogService는 새 로그 서비스 인스턴스를 생성합니다.
func NewLogService(repo repository.LogRepository) LogService {
	return &LogServiceImpl{
		repository: repo,
		log:        logger.GetLogger(),
	}
}

// SaveLogs는 로그를 저장합니다.
func (s *LogServiceImpl) SaveLogs(logs []domain.LogItem) error {
	return s.repository.SaveLogs(logs)
}

// QueryLogs는 로그를 쿼리합니다.
func (s *LogServiceImpl) QueryLogs(filter domain.LogFilter) (domain.LogQueryResult, error) {
	// 기본값 설정
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	
	// 100개 이상의 로그를 요청하는 경우 제한
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	
	return s.repository.QueryLogs(filter)
}

// GetServiceAggregation은 서비스 이름 집계를 가져옵니다.
func (s *LogServiceImpl) GetServiceAggregation(startTime, endTime int64) ([]domain.ServiceAggregation, error) {
	return s.repository.GetServiceAggregation(startTime, endTime)
}

// GetSeverityAggregation은 심각도 수준 집계를 가져옵니다.
func (s *LogServiceImpl) GetSeverityAggregation(startTime, endTime int64) ([]domain.SeverityAggregation, error) {
	return s.repository.GetSeverityAggregation(startTime, endTime)
}
