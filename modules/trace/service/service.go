package service

import (
	"time"

	"github.com/seongpil0948/otel-kafka-pg/modules/api/dto"
	"github.com/seongpil0948/otel-kafka-pg/modules/common/logger"
	"github.com/seongpil0948/otel-kafka-pg/modules/trace/domain"
	"github.com/seongpil0948/otel-kafka-pg/modules/trace/repository"
)

// TraceService는 트레이스 서비스 인터페이스입니다.
type TraceService interface {
	// 트레이스 저장
	SaveTraces(traces []domain.TraceItem) error

	// 특정 트레이스 조회
	GetTraceByID(traceID string) (*domain.Trace, error)

	// 트레이스 쿼리
	QueryTraces(filter domain.TraceFilter) (domain.TraceQueryResult, error)

	GetServices(startTime, endTime int64, filter string) (domain.ServiceListResult, error)

	GetServiceMetrics(startTime, endTime int64, serviceName string) ([]dto.ServiceMetric, error)
}

// TraceServiceImpl은 트레이스 서비스 구현체입니다.
type TraceServiceImpl struct {
	repository repository.TraceRepository
	log        logger.Logger
}

// NewTraceService는 새 트레이스 서비스 인스턴스를 생성합니다.
func NewTraceService(repo repository.TraceRepository) TraceService {
	return &TraceServiceImpl{
		repository: repo,
		log:        logger.GetLogger(),
	}
}

// SaveTraces는 트레이스를 저장합니다.
func (s *TraceServiceImpl) SaveTraces(traces []domain.TraceItem) error {
	return s.repository.SaveTraces(traces)
}

// GetTraceByID는 특정 트레이스를 조회합니다.
func (s *TraceServiceImpl) GetTraceByID(traceID string) (*domain.Trace, error) {
	return s.repository.GetTraceByID(traceID)
}

func (s *TraceServiceImpl) QueryTraces(filter domain.TraceFilter) (domain.TraceQueryResult, error) {
	now := time.Now().UnixMilli()
	if filter.EndTime == 0 {
		filter.EndTime = now
	}
	if filter.StartTime == 0 {
		filter.StartTime = now - 3600000 // 기본값: 최근 1시간
	}
	if filter.Limit == 0 {
		filter.Limit = 100 // 기본값: 100개
	}

	return s.repository.QueryTraces(filter)
}

func (s *TraceServiceImpl) GetServices(startTime, endTime int64, filter string) (domain.ServiceListResult, error) {
	return s.repository.GetServices(startTime, endTime, filter)
}

func (s *TraceServiceImpl) GetServiceMetrics(startTime, endTime int64, serviceName string) ([]dto.ServiceMetric, error) {
	return s.repository.GetServiceMetrics(startTime, endTime, serviceName)
}
