package telemetry

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/seongpil0948/otel-kafka-pg/modules/log/domain"
	_ "github.com/seongpil0948/otel-kafka-pg/modules/log/repository"
	_ "github.com/seongpil0948/otel-kafka-pg/modules/trace/domain"
	_ "github.com/seongpil0948/otel-kafka-pg/modules/trace/repository"
)

const (
	ToolQueryLogs     = "telemetry_query_logs"
	ToolGetSeverities = "telemetry_get_severities"
	ToolGetServices   = "telemetry_get_services"
	ToolQueryTraces   = "telemetry_query_traces"
	ToolGetTrace      = "telemetry_get_trace"
	ToolSearchErrors  = "telemetry_search_errors"
)

// isToolEnabledFunc는 도구가 활성화되어 있는지 확인하는 함수 타입입니다.
type isToolEnabledFunc func(string) bool

// RegisterTools는 텔레메트리 관련 도구들을 서버에 등록합니다.
func RegisterTools(
	server *server.MCPServer,
	isToolEnabled isToolEnabledFunc,
) error {
	// 로그 쿼리 도구 등록
	if isToolEnabled(ToolQueryLogs) {
		if err := registerQueryLogsTool(server); err != nil {
			return err
		}
	}

	// 심각도 조회 도구 등록
	if isToolEnabled(ToolGetSeverities) {
		if err := registerGetSeveritiesTool(server); err != nil {
			return err
		}
	}

	// 서비스 조회 도구 등록
	if isToolEnabled(ToolGetServices) {
		if err := registerGetServicesTool(server); err != nil {
			return err
		}
	}

	// 트레이스 쿼리 도구 등록
	if isToolEnabled(ToolQueryTraces) {
		if err := registerQueryTracesTool(server); err != nil {
			return err
		}
	}

	// 트레이스 조회 도구 등록
	if isToolEnabled(ToolGetTrace) {
		if err := registerGetTraceTool(server); err != nil {
			return err
		}
	}

	// 오류 검색 도구 등록
	if isToolEnabled(ToolSearchErrors) {
		if err := registerSearchErrorsTool(server); err != nil {
			return err
		}
	}

	return nil
}

// registerQueryLogsTool는 로그 쿼리 도구를 등록합니다.
func registerQueryLogsTool(server *server.MCPServer) error {
	// 로그 쿼리 도구 정의
	queryLogsTool := mcp.NewTool(ToolQueryLogs,
		mcp.WithDescription("시간, 서비스, 심각도 등으로 로그 조회"),
		mcp.WithNumber("start_time",
			mcp.Required(),
			mcp.Description("시작 시간 (Unix 타임스탬프, 밀리초)"),
		),
		mcp.WithNumber("end_time",
			mcp.Required(),
			mcp.Description("종료 시간 (Unix 타임스탬프, 밀리초)"),
		),
		mcp.WithString("service_name",
			mcp.Description("서비스 이름 필터"),
		),
		mcp.WithString("severity",
			mcp.Description("심각도 필터 (INFO, WARN, ERROR 등)"),
		),
		mcp.WithString("query",
			mcp.Description("검색어"),
		),
		mcp.WithNumber("limit",
			mcp.Description("결과 제한 수"),
		),
		mcp.WithNumber("offset",
			mcp.Description("결과 오프셋"),
		),
	)

	// 로그 쿼리 도구 핸들러 등록
	server.AddTool(queryLogsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// 요청 파라미터 파싱
		startTimeF, ok := request.Params.Arguments["start_time"].(float64)
		if !ok {
			return mcp.NewToolResultError("유효한 시작 시간이 필요합니다"), nil
		}
		startTime := int64(startTimeF)

		endTimeF, ok := request.Params.Arguments["end_time"].(float64)
		if !ok {
			return mcp.NewToolResultError("유효한 종료 시간이 필요합니다"), nil
		}
		endTime := int64(endTimeF)

		// 선택적 파라미터 처리
		var serviceName, severity, queryParam *string

		if serviceNameVal, ok := request.Params.Arguments["service_name"].(string); ok && serviceNameVal != "" {
			serviceName = &serviceNameVal
		}

		if severityVal, ok := request.Params.Arguments["severity"].(string); ok && severityVal != "" {
			severity = &severityVal
		}

		if queryVal, ok := request.Params.Arguments["query"].(string); ok && queryVal != "" {
			queryParam = &queryVal
		}

		limit := 20 // 기본값
		if limitVal, ok := request.Params.Arguments["limit"].(float64); ok {
			limit = int(limitVal)
		}

		offset := 0 // 기본값
		if offsetVal, ok := request.Params.Arguments["offset"].(float64); ok {
			offset = int(offsetVal)
		}

		// 이 예제에서는 실제 데이터베이스 쿼리 대신 샘플 데이터를 반환합니다.
		// 실제 구현에서는 로그 저장소를 사용하여 쿼리해야 합니다.

		// 샘플 로그 데이터
		sampleLogs := []domain.LogItem{
			{
				ID:          "log123",
				Timestamp:   time.Now().Add(-5*time.Minute).UnixNano() / 1000000,
				ServiceName: "api-gateway",
				Message:     "Request processed successfully",
				Severity:    "INFO",
				TraceID:     "trace123",
				SpanID:      "span123",
				Attributes: map[string]interface{}{
					"http.method":        "GET",
					"http.path":          "/api/v1/users",
					"http.status_code":   200,
					"http.response_time": 45.2,
				},
			},
			{
				ID:          "log124",
				Timestamp:   time.Now().Add(-3*time.Minute).UnixNano() / 1000000,
				ServiceName: "user-service",
				Message:     "Database connection timeout",
				Severity:    "ERROR",
				TraceID:     "trace124",
				SpanID:      "span124",
				Attributes: map[string]interface{}{
					"error.type":   "ConnectionTimeout",
					"db.system":    "postgresql",
					"db.operation": "query",
					"db.statement": "SELECT * FROM users WHERE id = ?",
				},
			},
		}

		// 결과 포맷팅
		result := fmt.Sprintf("로그 조회 결과 (총 %d개):\n\n", len(sampleLogs))

		// 쿼리 파라미터 요약
		result += "== 쿼리 파라미터 ==\n"
		result += fmt.Sprintf("시작 시간: %s\n", formatTimestamp(startTime))
		result += fmt.Sprintf("종료 시간: %s\n", formatTimestamp(endTime))

		if serviceName != nil {
			result += fmt.Sprintf("서비스: %s\n", *serviceName)
		}

		if severity != nil {
			result += fmt.Sprintf("심각도: %s\n", *severity)
		}

		if queryParam != nil {
			result += fmt.Sprintf("검색어: %s\n", *queryParam)
		}

		result += fmt.Sprintf("제한: %d, 오프셋: %d\n", limit, offset)
		result += "\n"

		// 로그 목록
		result += "== 로그 목록 ==\n"
		for i, log := range sampleLogs {
			result += fmt.Sprintf("%d. [%s] %s - %s\n",
				i+1,
				log.Severity,
				formatTimestamp(log.Timestamp),
				log.Message,
			)
			result += fmt.Sprintf("   서비스: %s\n", log.ServiceName)

			if log.TraceID != "" {
				result += fmt.Sprintf("   트레이스: %s\n", log.TraceID)
			}

			// 주요 속성 출력
			if len(log.Attributes) > 0 {
				result += "   주요 속성:\n"
				for k, v := range log.Attributes {
					result += fmt.Sprintf("     - %s: %v\n", k, v)
				}
			}

			result += "\n"
		}

		return mcp.NewToolResultText(result), nil
	})

	return nil
}

// registerGetSeveritiesTool는 심각도 조회 도구를 등록합니다.
func registerGetSeveritiesTool(server *server.MCPServer) error {
	// 심각도 조회 도구 정의
	getSeveritiesTool := mcp.NewTool(ToolGetSeverities,
		mcp.WithDescription("로그 심각도 수준별 개수 조회"),
		mcp.WithNumber("start_time",
			mcp.Required(),
			mcp.Description("시작 시간 (Unix 타임스탬프, 밀리초)"),
		),
		mcp.WithNumber("end_time",
			mcp.Required(),
			mcp.Description("종료 시간 (Unix 타임스탬프, 밀리초)"),
		),
	)

	// 심각도 조회 도구 핸들러 등록
	server.AddTool(getSeveritiesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// 요청 파라미터 파싱
		startTimeF, ok := request.Params.Arguments["start_time"].(float64)
		if !ok {
			return mcp.NewToolResultError("유효한 시작 시간이 필요합니다"), nil
		}
		startTime := int64(startTimeF)

		endTimeF, ok := request.Params.Arguments["end_time"].(float64)
		if !ok {
			return mcp.NewToolResultError("유효한 종료 시간이 필요합니다"), nil
		}
		endTime := int64(endTimeF)

		// 샘플 데이터
		severities := []domain.SeverityAggregation{
			{Name: "INFO", Count: 1250},
			{Name: "WARN", Count: 320},
			{Name: "ERROR", Count: 45},
			{Name: "FATAL", Count: 2},
		}

		// 결과 포맷팅
		result := fmt.Sprintf("로그 심각도 집계 결과 (%s ~ %s):\n\n",
			formatTimestamp(startTime),
			formatTimestamp(endTime),
		)

		// 총 로그 개수 계산
		totalLogs := 0
		for _, sev := range severities {
			totalLogs += sev.Count
		}

		result += fmt.Sprintf("총 로그 수: %d개\n\n", totalLogs)

		// 심각도별 개수 및 비율
		result += "심각도별 분포:\n"
		for _, sev := range severities {
			percentage := float64(sev.Count) / float64(totalLogs) * 100
			result += fmt.Sprintf("- %s: %d개 (%.1f%%)\n", sev.Name, sev.Count, percentage)
		}

		return mcp.NewToolResultText(result), nil
	})

	return nil
}

// registerGetServicesTool는 서비스 조회 도구를 등록합니다.
func registerGetServicesTool(server *server.MCPServer) error {
	// 서비스 조회 도구 정의
	getServicesTool := mcp.NewTool(ToolGetServices,
		mcp.WithDescription("로그 서비스별 개수 조회"),
		mcp.WithNumber("start_time",
			mcp.Required(),
			mcp.Description("시작 시간 (Unix 타임스탬프, 밀리초)"),
		),
		mcp.WithNumber("end_time",
			mcp.Required(),
			mcp.Description("종료 시간 (Unix 타임스탬프, 밀리초)"),
		),
	)

	// 서비스 조회 도구 핸들러 등록
	server.AddTool(getServicesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// 요청 파라미터 파싱
		startTimeF, ok := request.Params.Arguments["start_time"].(float64)
		if !ok {
			return mcp.NewToolResultError("유효한 시작 시간이 필요합니다"), nil
		}
		startTime := int64(startTimeF)

		endTimeF, ok := request.Params.Arguments["end_time"].(float64)
		if !ok {
			return mcp.NewToolResultError("유효한 종료 시간이 필요합니다"), nil
		}
		endTime := int64(endTimeF)

		// 샘플 데이터
		services := []domain.ServiceAggregation{
			{Name: "api-gateway", Count: 850},
			{Name: "user-service", Count: 420},
			{Name: "order-service", Count: 215},
			{Name: "payment-service", Count: 132},
		}

		// 결과 포맷팅
		result := fmt.Sprintf("서비스별 로그 집계 결과 (%s ~ %s):\n\n",
			formatTimestamp(startTime),
			formatTimestamp(endTime),
		)

		// 총 로그 개수 계산
		totalLogs := 0
		for _, svc := range services {
			totalLogs += svc.Count
		}

		result += fmt.Sprintf("총 로그 수: %d개\n\n", totalLogs)

		// 서비스별 개수 및 비율 (내림차순)
		result += "서비스별 분포 (로그 개수 내림차순):\n"
		for _, svc := range services {
			percentage := float64(svc.Count) / float64(totalLogs) * 100
			result += fmt.Sprintf("- %s: %d개 (%.1f%%)\n", svc.Name, svc.Count, percentage)
		}

		return mcp.NewToolResultText(result), nil
	})

	return nil
}

// registerQueryTracesTool는 트레이스 쿼리 도구를 등록합니다.
func registerQueryTracesTool(server *server.MCPServer) error {
	// 트레이스 쿼리 도구 정의
	queryTracesTool := mcp.NewTool(ToolQueryTraces,
		mcp.WithDescription("서비스, 기간, 상태 등으로 트레이스 조회"),
		mcp.WithNumber("start_time",
			mcp.Required(),
			mcp.Description("시작 시간 (Unix 타임스탬프, 밀리초)"),
		),
		mcp.WithNumber("end_time",
			mcp.Required(),
			mcp.Description("종료 시간 (Unix 타임스탬프, 밀리초)"),
		),
		mcp.WithString("service_name",
			mcp.Description("서비스 이름 필터"),
		),
		mcp.WithString("status",
			mcp.Description("상태 필터 (OK, ERROR)"),
		),
		mcp.WithNumber("min_duration",
			mcp.Description("최소 지속 시간 (밀리초)"),
		),
		mcp.WithNumber("max_duration",
			mcp.Description("최대 지속 시간 (밀리초)"),
		),
		mcp.WithString("query",
			mcp.Description("검색어"),
		),
		mcp.WithNumber("limit",
			mcp.Description("결과 제한 수"),
		),
		mcp.WithNumber("offset",
			mcp.Description("결과 오프셋"),
		),
	)

	// 트레이스 쿼리 도구 핸들러 등록
	server.AddTool(queryTracesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// 요청 파라미터 파싱
		startTimeF, ok := request.Params.Arguments["start_time"].(float64)
		if !ok {
			return mcp.NewToolResultError("유효한 시작 시간이 필요합니다"), nil
		}
		startTime := int64(startTimeF)

		endTimeF, ok := request.Params.Arguments["end_time"].(float64)
		if !ok {
			return mcp.NewToolResultError("유효한 종료 시간이 필요합니다"), nil
		}
		endTime := int64(endTimeF)

		// 선택적 파라미터 처리
		var serviceName, status, queryParam *string
		var minDuration, maxDuration *float64

		if serviceNameVal, ok := request.Params.Arguments["service_name"].(string); ok && serviceNameVal != "" {
			serviceName = &serviceNameVal
		}

		if statusVal, ok := request.Params.Arguments["status"].(string); ok && statusVal != "" {
			status = &statusVal
		}

		if minDurationVal, ok := request.Params.Arguments["min_duration"].(float64); ok {
			minDuration = &minDurationVal
		}

		if maxDurationVal, ok := request.Params.Arguments["max_duration"].(float64); ok {
			maxDuration = &maxDurationVal
		}

		if queryVal, ok := request.Params.Arguments["query"].(string); ok && queryVal != "" {
			queryParam = &queryVal
		}

		limit := 20 // 기본값
		if limitVal, ok := request.Params.Arguments["limit"].(float64); ok {
			limit = int(limitVal)
		}

		offset := 0 // 기본값
		if offsetVal, ok := request.Params.Arguments["offset"].(float64); ok {
			offset = int(offsetVal)
		}

		// 트레이스 그룹 데이터 (샘플)
		sampleTraceGroups := []struct {
			TraceID   string
			StartTime int64
			Duration  float64
			SpanCount int
			Services  []string
		}{
			{
				TraceID:   "trace123",
				StartTime: time.Now().Add(-10*time.Minute).UnixNano() / 1000000,
				Duration:  150.5,
				SpanCount: 2,
				Services:  []string{"api-gateway", "user-service"},
			},
			{
				TraceID:   "trace124",
				StartTime: time.Now().Add(-5*time.Minute).UnixNano() / 1000000,
				Duration:  320.8,
				SpanCount: 4,
				Services:  []string{"api-gateway", "order-service", "payment-service"},
			},
		}

		// 결과 포맷팅
		result := fmt.Sprintf("트레이스 조회 결과:\n\n")

		// 쿼리 파라미터 요약
		result += "== 쿼리 파라미터 ==\n"
		result += fmt.Sprintf("시작 시간: %s\n", formatTimestamp(startTime))
		result += fmt.Sprintf("종료 시간: %s\n", formatTimestamp(endTime))

		if serviceName != nil {
			result += fmt.Sprintf("서비스: %s\n", *serviceName)
		}

		if status != nil {
			result += fmt.Sprintf("상태: %s\n", *status)
		}

		if minDuration != nil {
			result += fmt.Sprintf("최소 지속 시간: %.2f ms\n", *minDuration)
		}

		if maxDuration != nil {
			result += fmt.Sprintf("최대 지속 시간: %.2f ms\n", *maxDuration)
		}

		if queryParam != nil {
			result += fmt.Sprintf("검색어: %s\n", *queryParam)
		}

		result += fmt.Sprintf("제한: %d, 오프셋: %d\n", limit, offset)
		result += "\n"

		// 트레이스 그룹 목록
		result += "== 트레이스 그룹 ==\n"
		for i, group := range sampleTraceGroups {
			result += fmt.Sprintf("%d. 트레이스 ID: %s\n", i+1, group.TraceID)
			result += fmt.Sprintf("   시작 시간: %s\n", formatTimestamp(group.StartTime))
			result += fmt.Sprintf("   지속 시간: %.2f ms\n", group.Duration)
			result += fmt.Sprintf("   스팬 개수: %d개\n", group.SpanCount)
			result += fmt.Sprintf("   서비스: %s\n", strings.Join(group.Services, ", "))
			result += "\n"
		}

		return mcp.NewToolResultText(result), nil
	})

	return nil
}

// registerGetTraceTool는 트레이스 조회 도구를 등록합니다.
func registerGetTraceTool(server *server.MCPServer) error {
	// 트레이스 조회 도구 정의
	getTraceTool := mcp.NewTool(ToolGetTrace,
		mcp.WithDescription("특정 트레이스 ID의 상세 정보 조회"),
		mcp.WithString("trace_id",
			mcp.Required(),
			mcp.Description("조회할 트레이스 ID"),
		),
	)

	// 트레이스 조회 도구 핸들러 등록
	server.AddTool(getTraceTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// 요청 파라미터 파싱
		traceID, ok := request.Params.Arguments["trace_id"].(string)
		if !ok || traceID == "" {
			return mcp.NewToolResultError("유효한 트레이스 ID가 필요합니다"), nil
		}

		// 샘플 스팬 데이터
		type Span struct {
			ID           string
			TraceID      string
			SpanID       string
			ParentSpanID string
			Name         string
			ServiceName  string
			StartTime    int64
			EndTime      int64
			Duration     float64
			Status       string
			Attributes   map[string]interface{}
		}

		sampleSpans := []Span{
			{
				ID:           "trace123-span1",
				TraceID:      "trace123",
				SpanID:       "span1",
				ParentSpanID: "",
				Name:         "GET /api/users",
				ServiceName:  "api-gateway",
				StartTime:    time.Now().Add(-10*time.Minute).UnixNano() / 1000000,
				EndTime:      time.Now().Add(-10*time.Minute).Add(150*time.Millisecond).UnixNano() / 1000000,
				Duration:     150.5,
				Status:       "OK",
				Attributes: map[string]interface{}{
					"http.method":      "GET",
					"http.url":         "/api/users",
					"http.status_code": 200,
				},
			},
			{
				ID:           "trace123-span2",
				TraceID:      "trace123",
				SpanID:       "span2",
				ParentSpanID: "span1",
				Name:         "FindUsers",
				ServiceName:  "user-service",
				StartTime:    time.Now().Add(-10*time.Minute).Add(20*time.Millisecond).UnixNano() / 1000000,
				EndTime:      time.Now().Add(-10*time.Minute).Add(120*time.Millisecond).UnixNano() / 1000000,
				Duration:     100,
				Status:       "OK",
				Attributes: map[string]interface{}{
					"db.system":    "postgresql",
					"db.operation": "query",
					"db.statement": "SELECT * FROM users LIMIT 100",
				},
			},
			{
				ID:           "trace123-span3",
				TraceID:      "trace123",
				SpanID:       "span3",
				ParentSpanID: "span2",
				Name:         "ExecuteSQL",
				ServiceName:  "user-service",
				StartTime:    time.Now().Add(-10*time.Minute).Add(40*time.Millisecond).UnixNano() / 1000000,
				EndTime:      time.Now().Add(-10*time.Minute).Add(100*time.Millisecond).UnixNano() / 1000000,
				Duration:     60,
				Status:       "OK",
				Attributes: map[string]interface{}{
					"db.system":    "postgresql",
					"db.instance":  "users-db",
					"db.statement": "SELECT * FROM users LIMIT 100",
				},
			},
		}

		// 결과 생성 - 샘플 트레이스 데이터 구성
		type Trace struct {
			TraceID   string
			Spans     []Span
			StartTime int64
			EndTime   int64
			Services  []string
			Total     int
		}

		trace := &Trace{
			TraceID:   traceID,
			Spans:     sampleSpans,
			StartTime: sampleSpans[0].StartTime,
			EndTime:   sampleSpans[0].EndTime,
			Services:  []string{"api-gateway", "user-service"},
			Total:     len(sampleSpans),
		}

		// 결과 포맷팅
		result := fmt.Sprintf("트레이스 상세 정보 (ID: %s):\n\n", traceID)
		result += fmt.Sprintf("시작 시간: %s\n", formatTimestamp(trace.StartTime))
		result += fmt.Sprintf("종료 시간: %s\n", formatTimestamp(trace.EndTime))
		result += fmt.Sprintf("총 지속 시간: %.2f ms\n", sampleSpans[0].Duration)
		result += fmt.Sprintf("서비스: %s\n", strings.Join(trace.Services, ", "))
		result += fmt.Sprintf("스팬 개수: %d개\n\n", trace.Total)

		// 스팬 목록
		result += "== 스팬 목록 ==\n"
		for i, span := range trace.Spans {
			indent := ""
			if span.ParentSpanID != "" {
				for p := span.ParentSpanID; p != ""; {
					indent += "  "
					// 실제 구현에서는 부모 스팬 ID를 기반으로 재귀적으로 깊이를 찾아야 함
					// 이 예제에서는 단순화하여 구현
					if p == "span1" {
						p = ""
					} else if p == "span2" {
						p = "span1"
					}
				}
			}

			result += fmt.Sprintf("%s%d. [%s] %s (%.2f ms)\n",
				indent,
				i+1,
				span.ServiceName,
				span.Name,
				span.Duration,
			)

			result += fmt.Sprintf("%s   시작: %s\n",
				indent,
				formatTimestampWithMicros(span.StartTime),
			)

			result += fmt.Sprintf("%s   상태: %s\n", indent, span.Status)

			// 주요 속성 출력
			if len(span.Attributes) > 0 {
				result += fmt.Sprintf("%s   주요 속성:\n", indent)
				for k, v := range span.Attributes {
					result += fmt.Sprintf("%s     - %s: %v\n", indent, k, v)
				}
			}

			result += "\n"
		}

		return mcp.NewToolResultText(result), nil
	})

	return nil
}

// registerSearchErrorsTool는 오류 검색 도구를 등록합니다.
func registerSearchErrorsTool(server *server.MCPServer) error {
	// 오류 검색 도구 정의
	searchErrorsTool := mcp.NewTool(ToolSearchErrors,
		mcp.WithDescription("오류 로그 및 실패한 트레이스 검색"),
		mcp.WithNumber("start_time",
			mcp.Required(),
			mcp.Description("시작 시간 (Unix 타임스탬프, 밀리초)"),
		),
		mcp.WithNumber("end_time",
			mcp.Required(),
			mcp.Description("종료 시간 (Unix 타임스탬프, 밀리초)"),
		),
		mcp.WithString("service_name",
			mcp.Description("서비스 이름 필터"),
		),
		mcp.WithString("error_type",
			mcp.Description("오류 유형 필터"),
		),
		mcp.WithString("query",
			mcp.Description("검색어"),
		),
		mcp.WithNumber("limit",
			mcp.Description("결과 제한 수"),
		),
	)

	// 오류 검색 도구 핸들러 등록
	server.AddTool(searchErrorsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// 요청 파라미터 파싱
		startTimeF, ok := request.Params.Arguments["start_time"].(float64)
		if !ok {
			return mcp.NewToolResultError("유효한 시작 시간이 필요합니다"), nil
		}
		startTime := int64(startTimeF)

		endTimeF, ok := request.Params.Arguments["end_time"].(float64)
		if !ok {
			return mcp.NewToolResultError("유효한 종료 시간이 필요합니다"), nil
		}
		endTime := int64(endTimeF)

		// 선택적 파라미터 처리
		var serviceName, errorType, queryParam *string

		if serviceNameVal, ok := request.Params.Arguments["service_name"].(string); ok && serviceNameVal != "" {
			serviceName = &serviceNameVal
		}

		if errorTypeVal, ok := request.Params.Arguments["error_type"].(string); ok && errorTypeVal != "" {
			errorType = &errorTypeVal
		}

		if queryVal, ok := request.Params.Arguments["query"].(string); ok && queryVal != "" {
			queryParam = &queryVal
		}

		limit := 20 // 기본값
		if limitVal, ok := request.Params.Arguments["limit"].(float64); ok {
			limit = int(limitVal)
		}

		// 샘플 오류 데이터
		sampleErrors := []struct {
			Timestamp   int64
			ServiceName string
			Message     string
			ErrorType   string
			TraceID     string
			StackTrace  string
		}{
			{
				Timestamp:   time.Now().Add(-3*time.Hour).UnixNano() / 1000000,
				ServiceName: "user-service",
				Message:     "Database connection timeout",
				ErrorType:   "ConnectionTimeout",
				TraceID:     "trace124",
				StackTrace:  "java.sql.SQLTimeoutException: Connection timeout\n  at org.postgresql.Driver.connect(...)\n  at com.example.UserService.findUser(...)",
			},
			{
				Timestamp:   time.Now().Add(-1*time.Hour).UnixNano() / 1000000,
				ServiceName: "payment-service",
				Message:     "Payment gateway returned error: Invalid card number",
				ErrorType:   "PaymentRejected",
				TraceID:     "trace125",
				StackTrace:  "com.example.PaymentException: Payment gateway returned error: Invalid card number\n  at com.example.PaymentService.processPayment(...)",
			},
		}

		// 결과 포맷팅
		result := fmt.Sprintf("오류 검색 결과 (총 %d개):\n\n", len(sampleErrors))

		// 쿼리 파라미터 요약
		result += "== 검색 파라미터 ==\n"
		result += fmt.Sprintf("시간 범위: %s ~ %s\n",
			formatTimestamp(startTime),
			formatTimestamp(endTime),
		)

		if serviceName != nil {
			result += fmt.Sprintf("서비스: %s\n", *serviceName)
		}

		if errorType != nil {
			result += fmt.Sprintf("오류 유형: %s\n", *errorType)
		}

		if queryParam != nil {
			result += fmt.Sprintf("검색어: %s\n", *queryParam)
		}

		result += fmt.Sprintf("제한: %d\n", limit)

		result += "\n"

		// 오류 목록
		result += "== 오류 목록 ==\n"
		for i, err := range sampleErrors {
			result += fmt.Sprintf("%d. [%s] %s\n",
				i+1,
				err.ServiceName,
				err.Message,
			)
			result += fmt.Sprintf("   시간: %s\n", formatTimestamp(err.Timestamp))
			result += fmt.Sprintf("   오류 유형: %s\n", err.ErrorType)
			result += fmt.Sprintf("   트레이스 ID: %s\n", err.TraceID)
			result += fmt.Sprintf("   스택 트레이스:\n%s\n",
				formatStackTrace(err.StackTrace, 5),
			)
			result += "\n"
		}

		return mcp.NewToolResultText(result), nil
	})

	return nil
}

// formatTimestamp는 Unix 타임스탬프를 읽기 쉬운 형식으로 변환합니다.
func formatTimestamp(timestamp int64) string {
	t := time.Unix(0, timestamp*int64(time.Millisecond))
	return t.Format("2006-01-02 15:04:05")
}

// formatTimestampWithMicros는 마이크로초까지 포함한 타임스탬프 형식을 반환합니다.
func formatTimestampWithMicros(timestamp int64) string {
	t := time.Unix(0, timestamp*int64(time.Millisecond))
	return t.Format("15:04:05.000000")
}

// formatStackTrace는 스택 트레이스를 최대 라인 수까지 포맷팅합니다.
func formatStackTrace(stackTrace string, maxLines int) string {
	lines := strings.Split(stackTrace, "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		lines = append(lines, "  ...")
	}

	result := ""
	for _, line := range lines {
		result += fmt.Sprintf("     %s\n", line)
	}

	return result
}
