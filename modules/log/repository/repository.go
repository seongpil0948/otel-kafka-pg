package repository

import (
	"fmt"
	"time"

	"github.com/seongpil0948/otel-kafka-pg/modules/common/db"
	"github.com/seongpil0948/otel-kafka-pg/modules/common/logger"
	"github.com/seongpil0948/otel-kafka-pg/modules/log/domain"
)

// LogRepository는 로그 저장소 인터페이스입니다.
type LogRepository interface {
	// 로그 저장
	SaveLogs(logs []domain.LogItem) error
	
	// 로그 쿼리
	QueryLogs(filter domain.LogFilter) (domain.LogQueryResult, error)
	
	// 로그 집계
	GetServiceAggregation(startTime, endTime int64) ([]domain.ServiceAggregation, error)
	GetSeverityAggregation(startTime, endTime int64) ([]domain.SeverityAggregation, error)
}

// PostgresLogRepository는 PostgreSQL 로그 저장소 구현체입니다.
type PostgresLogRepository struct {
	db  db.Database
	log logger.Logger
}

// NewLogRepository는 새 로그 저장소 인스턴스를 생성합니다.
func NewLogRepository(database db.Database) LogRepository {
	return &PostgresLogRepository{
		db:  database,
		log: logger.GetLogger(),
	}
}

// SaveLogs는 로그 데이터를 데이터베이스에 저장합니다.
func (r *PostgresLogRepository) SaveLogs(logs []domain.LogItem) error {
	if len(logs) == 0 {
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
			r.log.Error().Err(err).Msg("로그 저장 트랜잭션 롤백됨")
		}
	}()

	// 로그 데이터 저장
	for _, log := range logs {
		attributes, err := log.AttributesToJSON()
		if err != nil {
			r.log.Error().Err(err).Msg("Failed to convert log attributes to JSON")
			continue
		}

		_, err = tx.Exec(
			`INSERT INTO logs(
				id, timestamp, service_name, message, severity, 
				trace_id, span_id, attributes
			) VALUES($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (id) DO UPDATE SET
				service_name = EXCLUDED.service_name,
				message = EXCLUDED.message,
				severity = EXCLUDED.severity,
				trace_id = EXCLUDED.trace_id,
				span_id = EXCLUDED.span_id,
				attributes = EXCLUDED.attributes`,
			log.ID,
			log.Timestamp,
			log.ServiceName,
			log.Message,
			log.Severity,
			log.TraceID,
			log.SpanID,
			attributes,
		)

		if err != nil {
			return fmt.Errorf("failed to insert log: %w", err)
		}
	}

	// 트랜잭션 커밋
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.log.Info().Int("count", len(logs)).Msg("Successfully saved logs to database")
	return nil
}

// QueryLogs는 필터에 따라 로그를 쿼리합니다.
func (r *PostgresLogRepository) QueryLogs(filter domain.LogFilter) (domain.LogQueryResult, error) {
	startTime := time.Now()
	result := domain.LogQueryResult{
		Logs:       []domain.LogItem{},
		Services:   []domain.ServiceAggregation{},
		Severities: []domain.SeverityAggregation{},
		Total:      0,
		Took:       0,
	}

	// 쿼리 파라미터 배열
	queryParams := []interface{}{filter.StartTime, filter.EndTime}
	paramIndex := 3
	
	// 기본 WHERE 조건
	whereClause := "timestamp >= $1 AND timestamp <= $2"

	// 서비스명 필터
	if filter.ServiceName != nil && *filter.ServiceName != "" {
		whereClause += fmt.Sprintf(" AND service_name = $%d", paramIndex)
		queryParams = append(queryParams, *filter.ServiceName)
		paramIndex++
	}

	// 심각도 필터
	if filter.Severity != nil && *filter.Severity != "" {
		whereClause += fmt.Sprintf(" AND severity = $%d", paramIndex)
		queryParams = append(queryParams, *filter.Severity)
		paramIndex++
	}

	// 트레이스 연결 필터
	if filter.HasTrace {
		whereClause += " AND trace_id IS NOT NULL AND trace_id != ''"
	}

	// 검색어 필터
	if filter.Query != nil && *filter.Query != "" && *filter.Query != "*" {
		whereClause += fmt.Sprintf(` AND (
			message ILIKE $%d OR
			service_name ILIKE $%d
		)`, paramIndex, paramIndex)
		queryParams = append(queryParams, "%"+*filter.Query+"%")
		paramIndex++
	}

	// 1. 로그 조회 쿼리
	logsQuery := fmt.Sprintf(`
		SELECT 
			id,
			timestamp,
			service_name AS "serviceName",
			message,
			severity,
			trace_id AS "traceId",
			span_id AS "spanId",
			attributes
		FROM 
			logs
		WHERE 
			%s
		ORDER BY 
			timestamp DESC
		LIMIT $%d
		OFFSET $%d
	`, whereClause, paramIndex, paramIndex+1)

	queryParams = append(queryParams, filter.Limit, filter.Offset)

	// 2. 서비스 집계 쿼리
	servicesQuery := fmt.Sprintf(`
		SELECT 
			service_name AS name,
			COUNT(*) AS count
		FROM 
			logs
		WHERE 
			%s
		GROUP BY 
			service_name
		ORDER BY 
			count DESC
		LIMIT 20
	`, whereClause)

	// 3. 심각도 집계 쿼리
	severitiesQuery := fmt.Sprintf(`
		SELECT 
			severity AS name,
			COUNT(*) AS count
		FROM 
			logs
		WHERE 
			%s
		GROUP BY 
			severity
		ORDER BY 
			CASE 
				WHEN severity = 'FATAL' THEN 1
				WHEN severity = 'ERROR' THEN 2
				WHEN severity = 'WARN' THEN 3
				WHEN severity = 'INFO' THEN 4
				WHEN severity = 'DEBUG' THEN 5
				WHEN severity = 'TRACE' THEN 6
				ELSE 7
			END
	`, whereClause)

	// 4. 총 개수 카운트 쿼리
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) AS total
		FROM logs
		WHERE %s
	`, whereClause)

	// 로그 조회
	logsRows, err := r.db.Query(logsQuery, queryParams...)
	if err != nil {
		return result, fmt.Errorf("failed to query logs: %w", err)
	}
	defer logsRows.Close()

	for logsRows.Next() {
		var log domain.LogItem
		var attributesJSON string

		err := logsRows.Scan(
			&log.ID,
			&log.Timestamp,
			&log.ServiceName,
			&log.Message,
			&log.Severity,
			&log.TraceID,
			&log.SpanID,
			&attributesJSON,
		)
		if err != nil {
			return result, fmt.Errorf("failed to scan log row: %w", err)
		}

		if err := log.JSONToAttributes(attributesJSON); err != nil {
			r.log.Error().Err(err).Msg("failed to parse log attributes")
			log.Attributes = make(map[string]interface{})
		}

		result.Logs = append(result.Logs, log)
	}

	// 서비스 집계
	servicesRows, err := r.db.Query(servicesQuery, queryParams[:paramIndex-1]...)
	if err != nil {
		return result, fmt.Errorf("failed to query service aggregation: %w", err)
	}
	defer servicesRows.Close()

	for servicesRows.Next() {
		var service domain.ServiceAggregation
		if err := servicesRows.Scan(&service.Name, &service.Count); err != nil {
			return result, fmt.Errorf("failed to scan service aggregation row: %w", err)
		}
		result.Services = append(result.Services, service)
	}

	// 심각도 집계
	severitiesRows, err := r.db.Query(severitiesQuery, queryParams[:paramIndex-1]...)
	if err != nil {
		return result, fmt.Errorf("failed to query severity aggregation: %w", err)
	}
	defer severitiesRows.Close()

	for severitiesRows.Next() {
		var severity domain.SeverityAggregation
		if err := severitiesRows.Scan(&severity.Name, &severity.Count); err != nil {
			return result, fmt.Errorf("failed to scan severity aggregation row: %w", err)
		}
		result.Severities = append(result.Severities, severity)
	}

	// 총 개수 카운트
	var total int
	err = r.db.QueryRow(countQuery, queryParams[:paramIndex-1]...).Scan(&total)
	if err != nil {
		return result, fmt.Errorf("failed to count logs: %w", err)
	}
	result.Total = total

	// 실행 시간 계산
	result.Took = time.Since(startTime).Milliseconds()

	return result, nil
}

// GetServiceAggregation은 서비스 이름 집계를 가져옵니다.
func (r *PostgresLogRepository) GetServiceAggregation(startTime, endTime int64) ([]domain.ServiceAggregation, error) {
	query := `
		SELECT 
			service_name as name,
			COUNT(*) as count
		FROM logs
		WHERE timestamp >= $1 AND timestamp <= $2
		GROUP BY service_name
		ORDER BY count DESC
		LIMIT 20
	`
	
	rows, err := r.db.Query(query, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query service aggregation: %w", err)
	}
	defer rows.Close()

	var results []domain.ServiceAggregation
	for rows.Next() {
		var service domain.ServiceAggregation
		if err := rows.Scan(&service.Name, &service.Count); err != nil {
			return nil, fmt.Errorf("failed to scan service aggregation row: %w", err)
		}
		results = append(results, service)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating service aggregation rows: %w", err)
	}

	return results, nil
}

// GetSeverityAggregation은 심각도 수준 집계를 가져옵니다.
func (r *PostgresLogRepository) GetSeverityAggregation(startTime, endTime int64) ([]domain.SeverityAggregation, error) {
	query := `
		SELECT 
			severity as name,
			COUNT(*) as count
		FROM logs
		WHERE timestamp >= $1 AND timestamp <= $2
		GROUP BY severity
		ORDER BY 
			CASE 
				WHEN severity = 'FATAL' THEN 1
				WHEN severity = 'ERROR' THEN 2
				WHEN severity = 'WARN' THEN 3
				WHEN severity = 'INFO' THEN 4
				WHEN severity = 'DEBUG' THEN 5
				WHEN severity = 'TRACE' THEN 6
				ELSE 7
			END
	`
	
	rows, err := r.db.Query(query, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query severity aggregation: %w", err)
	}
	defer rows.Close()

	var results []domain.SeverityAggregation
	for rows.Next() {
		var severity domain.SeverityAggregation
		if err := rows.Scan(&severity.Name, &severity.Count); err != nil {
			return nil, fmt.Errorf("failed to scan severity aggregation row: %w", err)
		}
		results = append(results, severity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating severity aggregation rows: %w", err)
	}

	return results, nil
}
