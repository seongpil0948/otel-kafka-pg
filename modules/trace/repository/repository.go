package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/seongpil0948/otel-kafka-pg/modules/api/dto"
	"github.com/seongpil0948/otel-kafka-pg/modules/common/db"
	"github.com/seongpil0948/otel-kafka-pg/modules/common/logger"
	"github.com/seongpil0948/otel-kafka-pg/modules/trace/domain"
)

// TraceRepository는 트레이스 저장소 인터페이스입니다.
type TraceRepository interface {
	// 트레이스 저장
	SaveTraces(traces []domain.TraceItem) error

	// 특정 트레이스 조회
	GetTraceByID(traceID string) (*domain.Trace, error)

	// 트레이스 쿼리
	QueryTraces(filter domain.TraceFilter) (domain.TraceQueryResult, error)

	// 서비스 목록 조회
	GetServices(startTime, endTime int64, filter string) (domain.ServiceListResult, error)

	// 서비스 메트릭 조회
	GetServiceMetrics(startTime, endTime int64, serviceName string) ([]dto.ServiceMetric, error)
}

// PostgresTraceRepository는 PostgreSQL 트레이스 저장소 구현체입니다.
type PostgresTraceRepository struct {
	db  db.Database
	log logger.Logger
}

// NewTraceRepository는 새 트레이스 저장소 인스턴스를 생성합니다.
func NewTraceRepository(database db.Database) TraceRepository {
	return &PostgresTraceRepository{
		db:  database,
		log: logger.GetLogger(),
	}
}

// SaveTraces는 트레이스 데이터를 데이터베이스에 저장합니다.
func (r *PostgresTraceRepository) SaveTraces(traces []domain.TraceItem) error {
	if len(traces) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// 롤백 함수 준비
	defer func() {
		if err != nil {
			tx.Rollback()
			r.log.Error().Err(err).Msg("트레이스 저장 트랜잭션 롤백됨")
		}
	}()

	// 트레이스 데이터 저장
	for _, trace := range traces {
		attributes, err := trace.AttributesToJSON()
		if err != nil {
			r.log.Error().Err(err).Msg("Failed to convert trace attributes to JSON")
			continue
		}

		_, err = tx.Exec(
			`INSERT INTO traces(
				id, trace_id, span_id, parent_span_id, name, service_name, 
				start_time, end_time, duration, status, attributes
			) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			ON CONFLICT (id) DO UPDATE SET
				name = EXCLUDED.name,
				service_name = EXCLUDED.service_name,
				start_time = EXCLUDED.start_time,
				end_time = EXCLUDED.end_time,
				duration = EXCLUDED.duration,
				status = EXCLUDED.status,
				attributes = EXCLUDED.attributes`,
			trace.ID,
			trace.TraceID,
			trace.SpanID,
			trace.ParentSpanID,
			trace.Name,
			trace.ServiceName,
			trace.StartTime,
			trace.EndTime,
			trace.Duration,
			trace.Status,
			attributes,
		)

		if err != nil {
			return fmt.Errorf("failed to insert trace: %w", err)
		}
	}

	// 트랜잭션 커밋
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.log.Info().Int("count", len(traces)).Msg("Successfully saved traces to database")
	return nil
}

func (r *PostgresTraceRepository) GetTraceByID(traceID string) (*domain.Trace, error) {
	query := `
		SELECT 
			id, trace_id, span_id, parent_span_id,
			name, service_name, start_time, 
			end_time, duration, status, attributes
		FROM traces
		WHERE trace_id = $1
		ORDER BY start_time ASC
	`

	rows, err := r.db.Query(query, traceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query trace by ID: %w", err)
	}
	defer rows.Close()

	var spans []domain.Span
	var services = make(map[string]bool)
	var minStartTime int64 = 9223372036854775807 // int64 max
	var maxEndTime int64 = 0

	for rows.Next() {
		var span domain.Span
		var attributesJSON string
		var parentSpanID sql.NullString

		if err := rows.Scan(
			&span.ID,
			&span.TraceID,
			&span.SpanID,
			&parentSpanID,
			&span.Name,
			&span.ServiceName,
			&span.StartTime,
			&span.EndTime,
			&span.Duration,
			&span.Status,
			&attributesJSON,
		); err != nil {
			return nil, fmt.Errorf("failed to scan span row: %w", err)
		}

		// NULL 값 처리
		if parentSpanID.Valid {
			span.ParentSpanID = parentSpanID.String
		}

		// 속성 파싱
		if attributesJSON != "" {
			if err := json.Unmarshal([]byte(attributesJSON), &span.Attributes); err != nil {
				r.log.Error().Err(err).Msg("Failed to parse span attributes")
				span.Attributes = make(map[string]interface{})
			}
		} else {
			span.Attributes = make(map[string]interface{})
		}

		// 서비스 맵에 추가
		services[span.ServiceName] = true

		// 시작 및 종료 시간 추적
		if span.StartTime < minStartTime {
			minStartTime = span.StartTime
		}
		if span.EndTime > maxEndTime {
			maxEndTime = span.EndTime
		}

		spans = append(spans, span)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating span rows: %w", err)
	}

	if len(spans) == 0 {
		return nil, nil
	}

	// 서비스 목록 생성
	servicesList := make([]string, 0, len(services))
	for service := range services {
		servicesList = append(servicesList, service)
	}

	trace := &domain.Trace{
		TraceID:   traceID,
		Spans:     spans,
		StartTime: minStartTime,
		EndTime:   maxEndTime,
		Services:  servicesList,
		Total:     len(spans),
	}

	return trace, nil
}

func (r *PostgresTraceRepository) QueryTraces(filter domain.TraceFilter) (domain.TraceQueryResult, error) {
	startTime := time.Now()
	result := domain.TraceQueryResult{
		Traces:      []domain.TraceItem{},
		TraceGroups: []domain.TraceGroup{},
		Total:       0,
		Took:        0,
	}
	queryParams := []interface{}{filter.StartTime, filter.EndTime}
	paramIndex := 3
	whereClause := "start_time >= $1 AND start_time <= $2"

	sortFieldMap := map[string]string{
		"startTime":   "start_time",
		"endTime":     "end_time",
		"duration":    "duration",
		"name":        "name",
		"serviceName": "service_name",
		"status":      "status",
	}

	// 기본 정렬 설정
	sortField := "start_time"
	sortDirection := "DESC"

	// 클라이언트 정렬 필드를 DB 컬럼명으로 변환
	if dbField, ok := sortFieldMap[filter.SortField]; ok {
		sortField = dbField
	} else {
		// 지원하지 않는 정렬 필드인 경우 기본값 사용
		r.log.Warn().Str("sortField", filter.SortField).Msg("지원하지 않는 정렬 필드, 기본값(start_time) 사용")
	}

	if filter.SortDirection == "ASC" || filter.SortDirection == "DESC" {
		sortDirection = filter.SortDirection
	} else {
		// 지원하지 않는 정렬 방향인 경우 기본값 사용
		r.log.Warn().Str("sortDirection", filter.SortDirection).Msg("지원하지 않는 정렬 방향, 기본값(DESC) 사용")
	}

	// 서비스명 필터 수정
	if len(filter.ServiceNames) > 0 {
		placeholders := make([]string, len(filter.ServiceNames))
		for i := range filter.ServiceNames {
			placeholders[i] = fmt.Sprintf("$%d", paramIndex)
			queryParams = append(queryParams, filter.ServiceNames[i])
			paramIndex++
		}
		whereClause += fmt.Sprintf(" AND service_name IN (%s)", strings.Join(placeholders, ", "))
	}

	// RootSpansOnly 필터 추가
	if filter.RootSpansOnly {
		whereClause += " AND parent_span_id = '' OR parent_span_id IS NULL"
	}

	// 상태 필터
	if filter.Status != nil && *filter.Status != "" {
		whereClause += fmt.Sprintf(" AND status = $%d", paramIndex)
		queryParams = append(queryParams, *filter.Status)
		paramIndex++
	}

	// 지속 시간 필터
	if filter.MinDuration != nil {
		whereClause += fmt.Sprintf(" AND duration >= $%d", paramIndex)
		queryParams = append(queryParams, *filter.MinDuration)
		paramIndex++
	}

	if filter.MaxDuration != nil {
		whereClause += fmt.Sprintf(" AND duration <= $%d", paramIndex)
		queryParams = append(queryParams, *filter.MaxDuration)
		paramIndex++
	}

	// 검색어 필터
	if filter.Query != nil && *filter.Query != "" && *filter.Query != "*" {
		whereClause += fmt.Sprintf(` AND (
			name ILIKE $%d OR
			service_name ILIKE $%d OR
			trace_id ILIKE $%d
		)`, paramIndex, paramIndex, paramIndex)
		queryParams = append(queryParams, "%"+*filter.Query+"%")
		paramIndex++
	}
	if filter.RootSpansOnly {
		whereClause += " AND (parent_span_id = '' OR parent_span_id IS NULL)"
	}

	tracesQuery := fmt.Sprintf(`
        SELECT 
            id, trace_id AS "traceId", span_id AS "spanId", parent_span_id AS "parentSpanId",
            name, service_name AS "serviceName", start_time AS "startTime",
            end_time AS "endTime", duration, status, attributes
        FROM 
            traces
        WHERE 
            %s
        ORDER BY 
            %s %s
        LIMIT $%d
        OFFSET $%d
    `, whereClause, sortField, sortDirection, paramIndex, paramIndex+1)

	queryParams = append(queryParams, filter.Limit, filter.Offset)

	var groupSortField string
	if sortField == "start_time" {
		groupSortField = "MIN(start_time)"
	} else if sortField == "duration" {
		groupSortField = "MAX(end_time) - MIN(start_time)"
	} else {
		groupSortField = "trace_id"
	}

	traceGroupsQuery := fmt.Sprintf(`
		SELECT 
			trace_id AS "traceId",
			MIN(start_time) AS "startTime",
			MAX(end_time) - MIN(start_time) AS "duration",
			COUNT(*) AS "spanCount",
			json_agg(DISTINCT service_name) AS "services"
		FROM 
			traces
		WHERE 
			%s
		GROUP BY 
			trace_id
		ORDER BY 
			%s %s
		LIMIT 100
	`, whereClause, groupSortField, sortDirection)

	// 3. 총 개수 카운트 쿼리
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) AS total
		FROM traces
		WHERE %s
	`, whereClause)

	r.log.Debug().
		Str("sortField", filter.SortField).
		Str("dbSortField", sortField).
		Str("sortDirection", sortDirection).
		Str("whereClause", whereClause).
		Str("fullQuery", tracesQuery).
		Msg("정렬 적용된 트레이스 쿼리 실행")

	// 트레이스 조회
	tracesRows, err := r.db.Query(tracesQuery, queryParams...)
	if err != nil {
		return result, fmt.Errorf("failed to query traces: %w", err)
	}
	defer tracesRows.Close()

	for tracesRows.Next() {
		var trace domain.TraceItem
		var attributesJSON string
		var parentSpanID sql.NullString

		err := tracesRows.Scan(
			&trace.ID,
			&trace.TraceID,
			&trace.SpanID,
			&parentSpanID,
			&trace.Name,
			&trace.ServiceName,
			&trace.StartTime,
			&trace.EndTime,
			&trace.Duration,
			&trace.Status,
			&attributesJSON,
		)
		if err != nil {
			return result, fmt.Errorf("failed to scan trace row: %w", err)
		}

		// NULL 값 처리
		if parentSpanID.Valid {
			trace.ParentSpanID = parentSpanID.String
		}

		// 속성 파싱
		if attributesJSON != "" {
			if err := json.Unmarshal([]byte(attributesJSON), &trace.Attributes); err != nil {
				r.log.Error().Err(err).Msg("failed to parse trace attributes")
				trace.Attributes = make(map[string]interface{})
			}
		} else {
			trace.Attributes = make(map[string]interface{})
		}

		result.Traces = append(result.Traces, trace)
	}

	// 트레이스 그룹 조회
	traceGroupsRows, err := r.db.Query(traceGroupsQuery, queryParams[:paramIndex-1]...)
	if err != nil {
		return result, fmt.Errorf("failed to query trace groups: %w", err)
	}
	defer traceGroupsRows.Close()

	for traceGroupsRows.Next() {
		var traceGroup domain.TraceGroup
		var servicesJSON string

		if err := traceGroupsRows.Scan(
			&traceGroup.TraceID,
			&traceGroup.StartTime,
			&traceGroup.Duration,
			&traceGroup.SpanCount,
			&servicesJSON,
		); err != nil {
			return result, fmt.Errorf("failed to scan trace group row: %w", err)
		}

		// 서비스 목록 파싱
		var services []string
		if err := json.Unmarshal([]byte(servicesJSON), &services); err != nil {
			r.log.Error().Err(err).Msg("failed to parse services list")
			services = []string{}
		}
		traceGroup.Services = services

		result.TraceGroups = append(result.TraceGroups, traceGroup)
	}

	// 총 개수 카운트
	var total int
	err = r.db.QueryRow(countQuery, queryParams[:paramIndex-1]...).Scan(&total)
	if err != nil {
		return result, fmt.Errorf("failed to count traces: %w", err)
	}
	result.Total = total

	// 실행 시간 계산
	result.Took = time.Since(startTime).Milliseconds()

	return result, nil
}

// GetServices는 서비스 목록과 기본 통계 정보를 반환합니다.
func (r *PostgresTraceRepository) GetServices(startTime, endTime int64, filter string) (domain.ServiceListResult, error) {
	startQueryTime := time.Now()
	result := domain.ServiceListResult{
		Services: []domain.ServiceInfo{},
		Total:    0,
		Took:     0,
	}

	// 쿼리 파라미터 배열
	queryParams := []interface{}{startTime, endTime}
	paramIndex := 3

	// 기본 WHERE 조건
	whereClause := "start_time >= $1 AND start_time <= $2 AND service_name IS NOT NULL"

	// 서비스명 필터 (부분 일치 지원)
	if filter != "" {
		whereClause += fmt.Sprintf(" AND service_name ILIKE $%d", paramIndex)
		queryParams = append(queryParams, "%"+filter+"%")
		paramIndex++
	}

	// 서비스 조회 쿼리
	servicesQuery := fmt.Sprintf(`
        SELECT 
            service_name AS name,
            COUNT(*) AS count,
            COUNT(CASE WHEN status = 'ERROR' THEN 1 END) AS error_count,
            AVG(duration) AS avg_latency
        FROM 
            traces
        WHERE 
            %s
        GROUP BY 
            service_name
        ORDER BY 
            count DESC
        LIMIT 100
    `, whereClause)

	// 쿼리 실행
	rows, err := r.db.Query(servicesQuery, queryParams...)
	if err != nil {
		return result, fmt.Errorf("failed to query services: %w", err)
	}
	defer rows.Close()

	// 결과 처리
	for rows.Next() {
		var service domain.ServiceInfo
		var count, errorCount int
		var avgLatency float64

		if err := rows.Scan(&service.Name, &count, &errorCount, &avgLatency); err != nil {
			return result, fmt.Errorf("failed to scan service row: %w", err)
		}

		service.Count = count
		service.ErrorCount = errorCount
		service.AvgLatency = avgLatency

		// 오류율 계산
		if count > 0 {
			service.ErrorRate = float64(errorCount) / float64(count) * 100
		} else {
			service.ErrorRate = 0
		}

		result.Services = append(result.Services, service)
	}

	// 총 서비스 수는 반환된 서비스 목록의 길이
	result.Total = len(result.Services)

	// 실행 시간 계산
	result.Took = time.Since(startQueryTime).Milliseconds()

	return result, nil
}

// GetServiceMetrics 메서드 개선 (기존 메서드 수정)
func (r *PostgresTraceRepository) GetServiceMetrics(startTime, endTime int64, serviceName string) ([]dto.ServiceMetric, error) {
	// 쿼리 파라미터 배열
	queryParams := []interface{}{startTime, endTime}
	paramIndex := 3

	// 기본 WHERE 조건
	whereClause := "start_time >= $1 AND start_time <= $2 AND service_name IS NOT NULL"

	// 서비스명 필터 (선택적)
	if serviceName != "" {
		// 정확한 일치 대신 부분 일치로 수정
		whereClause += fmt.Sprintf(" AND service_name ILIKE $%d", paramIndex)
		queryParams = append(queryParams, "%"+serviceName+"%")
		paramIndex++
	}

	query := fmt.Sprintf(`
        SELECT 
            service_name AS name,
            COUNT(*) AS count,
            COUNT(CASE WHEN status = 'ERROR' THEN 1 END) AS error_count,
            AVG(duration) AS avg_latency,
            PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY duration) AS p95_latency,
            PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY duration) AS p99_latency
        FROM 
            traces
        WHERE 
            %s
        GROUP BY 
            service_name
        ORDER BY 
            count DESC
        LIMIT 50
    `, whereClause)

	rows, err := r.db.Query(query, queryParams...)
	if err != nil {
		return nil, fmt.Errorf("failed to query service metrics: %w", err)
	}
	defer rows.Close()

	var services []dto.ServiceMetric
	for rows.Next() {
		var service dto.ServiceMetric
		var count, errorCount int64
		var avgLatency, p95Latency, p99Latency sql.NullFloat64

		err := rows.Scan(&service.Name, &count, &errorCount, &avgLatency, &p95Latency, &p99Latency)
		if err != nil {
			return nil, fmt.Errorf("failed to scan service metrics row: %w", err)
		}

		service.RequestCount = count
		service.ErrorCount = errorCount

		if avgLatency.Valid {
			service.AvgLatency = avgLatency.Float64
		}

		if p95Latency.Valid {
			service.P95Latency = p95Latency.Float64
		}

		if p99Latency.Valid {
			service.P99Latency = p99Latency.Float64
		}

		service.ErrorRate = 0
		if count > 0 {
			service.ErrorRate = float64(errorCount) / float64(count) * 100
		}

		services = append(services, service)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating service metrics rows: %w", err)
	}

	return services, nil
}
