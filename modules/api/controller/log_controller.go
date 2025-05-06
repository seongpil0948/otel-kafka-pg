package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/seongpil0948/otel-kafka-pg/modules/api/dto"
	"github.com/seongpil0948/otel-kafka-pg/modules/common/logger"
	"github.com/seongpil0948/otel-kafka-pg/modules/log/domain"
	"github.com/seongpil0948/otel-kafka-pg/modules/log/service"
)

// LogController는 로그 관련 API 핸들러를 관리합니다
type LogController struct {
	logService service.LogService
	logger     logger.Logger
}

// NewLogController는 새 로그 컨트롤러를 생성합니다
func NewLogController(logService service.LogService, logger logger.Logger) *LogController {
	return &LogController{
		logService: logService,
		logger:     logger,
	}
}

// QueryLogs godoc
//
//	@Summary		로그 목록 조회
//	@Description	필터 조건에 맞는 로그 목록을 조회합니다
//	@Tags			logs
//	@Accept			json
//	@Produce		json
//	@Param			startTime	query		int		false	"시작 시간 (밀리초 타임스탬프)"
//	@Param			endTime		query		int		false	"종료 시간 (밀리초 타임스탬프)"
//	@Param			serviceNames	query		[]string	false	"서비스 이름 목록"
//	@Param			severity	query		string	false	"심각도 (INFO, WARN, ERROR, FATAL 등)"
//	@Param			hasTrace	query		boolean	false	"트레이스 ID가 있는 로그만 필터링"
//	@Param			query		query		string	false	"검색어"
//	@Param			limit		query		int		false	"한 페이지당 항목 수"	default(20)
//	@Param			offset		query		int		false	"오프셋"			default(0)
//	@Success		200			{object}	dto.Response{data=dto.LogsResponse}
//	@Failure		400			{object}	dto.Response
//	@Failure		500			{object}	dto.Response
//	@Router			/logs [get]
func (c *LogController) QueryLogs(ctx *gin.Context) {
	// 요청 매개변수 파싱
	var params dto.LogFilterParams
	if err := ctx.ShouldBindQuery(&params); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    http.StatusBadRequest,
				Message: "잘못된 요청 매개변수: " + err.Error(),
			},
		})
		return
	}

	// 기본 시간 범위 설정 (기본값: 최근 1시간)
	now := time.Now().UnixMilli()
	if params.EndTime == 0 {
		params.EndTime = now
	}
	if params.StartTime == 0 {
		params.StartTime = now - 3600000 // 1시간 전
	}

	// 기본 페이지네이션 설정
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	// 로그 필터 구성
	filter := domain.LogFilter{
		StartTime: params.StartTime,
		EndTime:   params.EndTime,
		HasTrace:  params.HasTrace,
		Limit:     params.Limit,
		Offset:    params.Offset,
	}

	if len(params.ServiceNames) > 0 {
		filter.ServiceNames = params.ServiceNames
	}
	if params.Severity != "" {
		filter.Severity = &params.Severity
	}
	if params.Query != "" {
		filter.Query = &params.Query
	}

	// 쿼리 실행
	result, err := c.logService.QueryLogs(filter)
	if err != nil {
		c.logger.Error().Err(err).Msg("로그 쿼리 실패")
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    http.StatusInternalServerError,
				Message: "로그 조회 중 오류가 발생했습니다",
			},
		})
		return
	}

	// 심각도 및 서비스 목록 추출
	severities := make(map[string]bool)
	services := make(map[string]bool)

	for _, log := range result.Logs {
		if log.Severity != "" {
			severities[log.Severity] = true
		}
		if log.ServiceName != "" {
			services[log.ServiceName] = true
		}
	}

	severitiesList := make([]string, 0, len(severities))
	for severity := range severities {
		severitiesList = append(severitiesList, severity)
	}

	servicesList := make([]string, 0, len(services))
	for service := range services {
		servicesList = append(servicesList, service)
	}

	// 응답 구성
	response := dto.LogsResponse{
		Logs: result.Logs,
		Pagination: dto.Pagination{
			Total:  result.Total,
			Limit:  params.Limit,
			Offset: params.Offset,
		},
		TimeRange: dto.TimeRange{
			StartTime: params.StartTime,
			EndTime:   params.EndTime,
		},
		Severities: severitiesList,
		Services:   servicesList,
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data:    response,
	})
}

// GetLogsByTraceID godoc
//
//	@Summary		트레이스 ID로 관련 로그 조회
//	@Description	특정 트레이스 ID와 관련된 로그를 조회합니다
//	@Tags			logs
//	@Accept			json
//	@Produce		json
//	@Param			traceId	path		string	true	"Trace ID"
//	@Success		200		{object}	dto.Response{data=dto.LogsResponse}
//	@Failure		400		{object}	dto.Response
//	@Failure		500		{object}	dto.Response
//	@Router			/logs/trace/{traceId} [get]
func (c *LogController) GetLogsByTraceID(ctx *gin.Context) {
	traceID := ctx.Param("traceId")
	if traceID == "" {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    http.StatusBadRequest,
				Message: "트레이스 ID가 필요합니다",
			},
		})
		return
	}

	// 페이지네이션 파라미터 파싱
	limitStr := ctx.DefaultQuery("limit", "20")
	offsetStr := ctx.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// 시간 범위 파싱 (선택적)
	startTimeStr := ctx.DefaultQuery("startTime", "")
	endTimeStr := ctx.DefaultQuery("endTime", "")

	// 기본 시간 범위 설정 (기본값: 최근 24시간)
	now := time.Now().UnixMilli()
	startTime := now - 86400000 // 24시간 전
	endTime := now

	if startTimeStr != "" {
		if parsedTime, err := strconv.ParseInt(startTimeStr, 10, 64); err == nil {
			startTime = parsedTime
		}
	}

	if endTimeStr != "" {
		if parsedTime, err := strconv.ParseInt(endTimeStr, 10, 64); err == nil {
			endTime = parsedTime
		}
	}

	// 필터 설정
	filter := domain.LogFilter{
		StartTime: startTime,
		EndTime:   endTime,
		HasTrace:  true,
		Limit:     limit,
		Offset:    offset,
		Query:     &traceID, // 트레이스 ID로 검색
	}

	// 쿼리 실행
	result, err := c.logService.QueryLogs(filter)
	if err != nil {
		c.logger.Error().Err(err).Str("traceId", traceID).Msg("트레이스 관련 로그 쿼리 실패")
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    http.StatusInternalServerError,
				Message: "로그 조회 중 오류가 발생했습니다",
			},
		})
		return
	}

	// 응답 구성
	response := dto.LogsResponse{
		Logs: result.Logs,
		Pagination: dto.Pagination{
			Total:  result.Total,
			Limit:  limit,
			Offset: offset,
		},
		TimeRange: dto.TimeRange{
			StartTime: startTime,
			EndTime:   endTime,
		},
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data:    response,
	})
}

// GetLogSummary godoc
//
//	@Summary		로그 요약 정보 조회
//	@Description	특정 기간의 로그에 대한 요약 정보를 조회합니다
//	@Tags			logs
//	@Accept			json
//	@Produce		json
//	@Param			startTime	query		int		false	"시작 시간 (밀리초 타임스탬프)"
//	@Param			endTime		query		int		false	"종료 시간 (밀리초 타임스탬프)"
//	@Param			serviceName	query		string	false	"서비스 이름(선택적)"
//	@Success		200			{object}	dto.Response
//	@Failure		400			{object}	dto.Response
//	@Failure		500			{object}	dto.Response
//	@Router			/logs/summary [get]
func (c *LogController) GetLogSummary(ctx *gin.Context) {
	// 시간 범위 파싱
	startTimeStr := ctx.DefaultQuery("startTime", "")
	endTimeStr := ctx.DefaultQuery("endTime", "")

	// 기본 시간 범위 설정 (기본값: 최근 1시간)
	now := time.Now().UnixMilli()
	startTime := now - 3600000 // 1시간 전
	endTime := now

	if startTimeStr != "" {
		if parsedTime, err := strconv.ParseInt(startTimeStr, 10, 64); err == nil {
			startTime = parsedTime
		}
	}

	if endTimeStr != "" {
		if parsedTime, err := strconv.ParseInt(endTimeStr, 10, 64); err == nil {
			endTime = parsedTime
		}
	}

	// 서비스별 및 심각도별 집계를 가져옵니다
	var serviceAggs []domain.ServiceAggregation
	var severityAggs []domain.SeverityAggregation
	var err error

	serviceAggs, err = c.logService.GetServiceAggregation(startTime, endTime)
	if err != nil {
		c.logger.Error().Err(err).Msg("서비스 집계 실패")
	}

	severityAggs, err = c.logService.GetSeverityAggregation(startTime, endTime)
	if err != nil {
		c.logger.Error().Err(err).Msg("심각도 집계 실패")
	}
	// 오류 발생 시 빈 배열로 설정
	if serviceAggs == nil {
		serviceAggs = []domain.ServiceAggregation{}
	}
	if severityAggs == nil {
		severityAggs = []domain.SeverityAggregation{}
	}

	// 응답 구성
	summary := map[string]interface{}{
		"timeRange": dto.TimeRange{
			StartTime: startTime,
			EndTime:   endTime,
		},
		"services":   serviceAggs,
		"severities": severityAggs,
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data:    summary,
	})
}
