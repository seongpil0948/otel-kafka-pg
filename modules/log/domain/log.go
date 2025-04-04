package domain

import (
	"encoding/json"
)

// LogItem은 로그 도메인 엔티티입니다.
type LogItem struct {
	ID          string                 `json:"id"`
	Timestamp   int64                  `json:"timestamp"`
	ServiceName string                 `json:"serviceName"`
	Message     string                 `json:"message"`
	Severity    string                 `json:"severity"`
	TraceID     string                 `json:"traceId,omitempty"`
	SpanID      string                 `json:"spanId,omitempty"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
}

// SeverityLevel은 로그 심각도 수준을 정의합니다.
type SeverityLevel int

const (
	TraceSeverity SeverityLevel = iota + 1
	DebugSeverity
	InfoSeverity
	WarnSeverity
	ErrorSeverity
	FatalSeverity
)

// SeverityNumberToText는 심각도 번호를 텍스트로 변환합니다.
func SeverityNumberToText(severityNumber int) string {
	severityMap := map[int]string{
		1:  "TRACE",
		5:  "DEBUG",
		9:  "INFO",
		13: "WARN",
		17: "ERROR",
		21: "FATAL",
	}

	if severity, exists := severityMap[severityNumber]; exists {
		return severity
	}
	return "INFO"
}

// AttributesToJSON은 속성 맵을 JSON 문자열로 변환합니다.
func (l *LogItem) AttributesToJSON() ([]byte, error) {
	if l.Attributes == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(l.Attributes)
}

// JSONToAttributes는 JSON 문자열을 속성 맵으로 변환합니다.
func (l *LogItem) JSONToAttributes(jsonStr string) error {
	if jsonStr == "" {
		l.Attributes = make(map[string]interface{})
		return nil
	}
	return json.Unmarshal([]byte(jsonStr), &l.Attributes)
}

// ServiceAggregation은 서비스별 로그 개수를 나타냅니다.
type ServiceAggregation struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// SeverityAggregation은 심각도별 로그 개수를 나타냅니다.
type SeverityAggregation struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// LogFilter는 로그 필터링 옵션을 정의합니다.
type LogFilter struct {
	StartTime   int64   `json:"startTime"`
	EndTime     int64   `json:"endTime"`
	ServiceName *string `json:"serviceName,omitempty"`
	Severity    *string `json:"severity,omitempty"`
	HasTrace    bool    `json:"hasTrace"`
	Query       *string `json:"query,omitempty"`
	Limit       int     `json:"limit"`
	Offset      int     `json:"offset"`
}

// LogQueryResult는 로그 쿼리 결과를 정의합니다.
type LogQueryResult struct {
	Logs       []LogItem            `json:"logs"`
	Services   []ServiceAggregation `json:"services"`
	Severities []SeverityAggregation `json:"severities"`
	Total      int                  `json:"total"`
	Took       int64                `json:"took"`
}
