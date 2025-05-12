package dto

import (
	"github.com/seongpil0948/otel-kafka-pg/modules/log/domain"
	traceDomain "github.com/seongpil0948/otel-kafka-pg/modules/trace/domain"
)

// Response 기본 응답 구조체
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo 오류 정보 구조체
type ErrorInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Pagination 페이지네이션 정보
type Pagination struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// TimeRange 시간 범위 정보
type TimeRange struct {
	StartTime int64 `json:"startTime"`
	EndTime   int64 `json:"endTime"`
}

// TracesResponse 트레이스 목록 응답
type TracesResponse struct {
	Traces        []traceDomain.TraceItem `json:"traces"`
	Pagination    Pagination              `json:"pagination"`
	TimeRange     TimeRange               `json:"timeRange"`
	Services      []string                `json:"services,omitempty"`
	TotalDuration int64                   `json:"totalDuration,omitempty"`
	SortField     string                  `json:"sortField,omitempty"`
	SortDirection string                  `json:"sortDirection,omitempty"`
}

// TraceDetailResponse 트레이스 상세 응답
type TraceDetailResponse struct {
	Trace       *traceDomain.Trace `json:"trace"`
	RelatedLogs []domain.LogItem   `json:"relatedLogs,omitempty"`
}

// LogsResponse 로그 목록 응답
type LogsResponse struct {
	Logs       []domain.LogItem `json:"logs"`
	Pagination Pagination       `json:"pagination"`
	TimeRange  TimeRange        `json:"timeRange"`
	Severities []string         `json:"severities,omitempty"`
	Services   []string         `json:"services,omitempty"`
}

// ServiceMetricsResponse 서비스 지표 응답
type ServiceMetricsResponse struct {
	Services        []ServiceMetric `json:"services"`
	TimeRange       TimeRange       `json:"timeRange"`
	TotalRequests   int64           `json:"totalRequests"`
	TotalErrors     int64           `json:"totalErrors"`
	AvgLatency      float64         `json:"avgLatency"`
	ErrorPercentage float64         `json:"errorPercentage"`
}

// ServiceMetric 서비스 지표 정보
type ServiceMetric struct {
	Name         string  `json:"name"`
	RequestCount int64   `json:"requestCount"`
	ErrorCount   int64   `json:"errorCount"`
	AvgLatency   float64 `json:"avgLatency"`
	P95Latency   float64 `json:"p95Latency,omitempty"`
	P99Latency   float64 `json:"p99Latency,omitempty"`
	ErrorRate    float64 `json:"errorRate"`
}

// LogFilter 로그 필터링 매개변수
type LogFilterParams struct {
	StartTime     int64    `form:"startTime"`
	EndTime       int64    `form:"endTime"`
	ServiceNames  []string `form:"serviceName"`
	Severity      string   `form:"severity"`
	HasTrace      bool     `form:"hasTrace"`
	Query         string   `form:"query"`
	Limit         int      `form:"limit,default=20"`
	Offset        int      `form:"offset,default=0"`
	RootSpansOnly bool     `form:"rootSpansOnly"` // 루트 스팬만 필터링 옵션 추가
}

// TraceFilterParams 트레이스 필터링 매개변수
type TraceFilterParams struct {
	StartTime     int64    `form:"startTime"`
	EndTime       int64    `form:"endTime"`
	ServiceNames  []string `form:"serviceName"`
	Status        string   `form:"status"`
	MinDuration   *int64   `form:"minDuration"`
	MaxDuration   *int64   `form:"maxDuration"`
	Query         string   `form:"query"`
	Limit         int      `form:"limit,default=20"`
	Offset        int      `form:"offset,default=0"`
	RootSpansOnly bool     `form:"rootSpansOnly"` // 루트 스팬만 필터링 옵션 추가
	SortField     string   `form:"sortField"`
	SortDirection string   `form:"sortDirection"`
}
