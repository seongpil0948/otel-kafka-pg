package domain

import (
	"encoding/json"
)

// TraceItem은 트레이스 데이터 구조를 정의합니다.
type TraceItem struct {
	ID           string                 `json:"id"`
	TraceID      string                 `json:"traceId"`
	SpanID       string                 `json:"spanId"`
	ParentSpanID string                 `json:"parentSpanId,omitempty"`
	Name         string                 `json:"name"`
	ServiceName  string                 `json:"serviceName"`
	StartTime    int64                  `json:"startTime"`
	EndTime      int64                  `json:"endTime"`
	Duration     float64                `json:"duration"`
	Status       string                 `json:"status,omitempty"`
	Attributes   map[string]interface{} `json:"attributes,omitempty"`
}

// Span은 스팬 데이터 구조를 정의합니다.
type Span struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	ServiceName  string                 `json:"serviceName"`
	StartTime    int64                  `json:"startTime"`
	EndTime      int64                  `json:"endTime"`
	Duration     float64                `json:"duration"`
	ParentSpanID string                 `json:"parentSpanId,omitempty"`
	TraceID      string                 `json:"traceId"`
	SpanID       string                 `json:"spanId"`
	Status       string                 `json:"status,omitempty"`
	Attributes   map[string]interface{} `json:"attributes,omitempty"`
}

// Trace는 전체 트레이스 데이터 구조를 정의합니다.
type Trace struct {
	TraceID   string   `json:"traceId"`
	Spans     []Span   `json:"spans"`
	StartTime int64    `json:"startTime"`
	EndTime   int64    `json:"endTime"`
	Services  []string `json:"services"`
	Total     int      `json:"total"`
}

// TraceGroup은 트레이스 그룹 데이터 구조를 정의합니다.
type TraceGroup struct {
	TraceID   string   `json:"traceId"`
	StartTime int64    `json:"startTime"`
	Duration  float64  `json:"duration"`
	SpanCount int      `json:"spanCount"`
	Services  []string `json:"services"`
}

// AttributesToJSON은 속성 맵을 JSON 문자열로 변환합니다.
func (t *TraceItem) AttributesToJSON() ([]byte, error) {
	if t.Attributes == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(t.Attributes)
}

// JSONToAttributes는 JSON 문자열을 속성 맵으로 변환합니다.
func (t *TraceItem) JSONToAttributes(jsonStr string) error {
	if jsonStr == "" {
		t.Attributes = make(map[string]interface{})
		return nil
	}
	return json.Unmarshal([]byte(jsonStr), &t.Attributes)
}

// AttributesToJSON은 속성 맵을 JSON 문자열로 변환합니다. (Span용)
func (s *Span) AttributesToJSON() ([]byte, error) {
	if s.Attributes == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(s.Attributes)
}

// JSONToAttributes는 JSON 문자열을 속성 맵으로 변환합니다. (Span용)
func (s *Span) JSONToAttributes(jsonStr string) error {
	if jsonStr == "" {
		s.Attributes = make(map[string]interface{})
		return nil
	}
	return json.Unmarshal([]byte(jsonStr), &s.Attributes)
}

// TraceFilter는 트레이스 필터링 옵션을 정의합니다.
type TraceFilter struct {
	StartTime     int64    `json:"startTime"`
	EndTime       int64    `json:"endTime"`
	ServiceNames  []string `json:"serviceNames,omitempty"`
	Status        *string  `json:"status,omitempty"`
	MinDuration   *float64 `json:"minDuration,omitempty"`
	MaxDuration   *float64 `json:"maxDuration,omitempty"`
	Query         *string  `json:"query,omitempty"`
	Limit         int      `json:"limit"`
	Offset        int      `json:"offset"`
	RootSpansOnly bool     `json:"rootSpansOnly,omitempty"`
}

// TraceQueryResult는 트레이스 쿼리 결과를 정의합니다.
type TraceQueryResult struct {
	Traces      []TraceItem  `json:"traces"`
	TraceGroups []TraceGroup `json:"traceGroups"`
	Total       int          `json:"total"`
	Took        int64        `json:"took"`
}

type ServiceInfo struct {
	Name       string  `json:"name"`
	Count      int     `json:"count"`
	ErrorCount int     `json:"errorCount"`
	ErrorRate  float64 `json:"errorRate"`
	AvgLatency float64 `json:"avgLatency"`
}

// ServiceListResult는 서비스 목록 쿼리 결과를 정의합니다.
type ServiceListResult struct {
	Services []ServiceInfo `json:"services"`
	Total    int           `json:"total"`
	Took     int64         `json:"took"`
}
