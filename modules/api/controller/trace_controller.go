package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/seongpil0948/otel-kafka-pg/modules/api/dto"
	"github.com/seongpil0948/otel-kafka-pg/modules/common/logger"
	traceDomain "github.com/seongpil0948/otel-kafka-pg/modules/trace/domain"
	"github.com/seongpil0948/otel-kafka-pg/modules/trace/service"
)

// TraceController는 트레이스 관련 API 핸들러를 관리합니다
type TraceController struct {
	traceService service.TraceService
	logger       logger.Logger
}

// NewTraceController는 새 트레이스 컨트롤러를 생성합니다
func NewTraceController(traceService service.TraceService, logger logger.Logger) *TraceController {
	return &TraceController{
		traceService: traceService,
		logger:       logger,
	}
}

// GetTraceByID godoc
//	@Summary		트레이스 ID로 상세 정보 조회
//	@Description	특정 트레이스 ID에 대한 상세 정보를 조회합니다
//	@Tags			traces
//	@Accept			json
//	@Produce		json
//	@Param			traceId	path		string	true	"Trace ID"
//	@Success		200		{object}	dto.Response{data=dto.TraceDetailResponse}
//	@Failure		400		{object}	dto.Response
//	@Failure		404		{object}	dto.Response
//	@Failure		500		{object}	dto.Response
//	@Router			/traces/{traceId} [get]
func (c *TraceController) GetTraceByID(ctx *gin.Context) {
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

	trace, err := c.traceService.GetTraceByID(traceID)
	if err != nil {
		c.logger.Error().Err(err).Str("traceId", traceID).Msg("트레이스 조회 실패")

		status := http.StatusInternalServerError
		message := "트레이스 조회 중 오류가 발생했습니다"

		if err.Error() == "sql: no rows in result set" {
			status = http.StatusNotFound
			message = "트레이스를 찾을 수 없습니다"
		}

		ctx.JSON(status, dto.Response{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    status,
				Message: message,
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data: dto.TraceDetailResponse{
			Trace: trace,
		},
	})
}

// QueryTraces godoc
//	@Summary		트레이스 목록 조회
//	@Description	필터 조건에 맞는 트레이스 목록을 조회합니다
//	@Tags			traces
//	@Accept			json
//	@Produce		json
//	@Param			startTime	query		int		false	"시작 시간 (밀리초 타임스탬프)"
//	@Param			endTime		query		int		false	"종료 시간 (밀리초 타임스탬프)"
//	@Param			serviceName	query		string	false	"서비스 이름"
//	@Param			status		query		string	false	"상태 (OK, ERROR 등)"
//	@Param			minDuration	query		int		false	"최소 지속 시간 (밀리초)"
//	@Param			maxDuration	query		int		false	"최대 지속 시간 (밀리초)"
//	@Param			query		query		string	false	"검색어"
//	@Param			limit		query		int		false	"한 페이지당 항목 수"	default(20)
//	@Param			offset		query		int		false	"오프셋"			default(0)
//	@Success		200			{object}	dto.Response{data=dto.TracesResponse}
//	@Failure		400			{object}	dto.Response
//	@Failure		500			{object}	dto.Response
//	@Router			/traces [get]
func (c *TraceController) QueryTraces(ctx *gin.Context) {
	// 시간 범위 설정
	var params dto.TraceFilterParams
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

	// 트레이스 필터 구성
	filter := traceDomain.TraceFilter{
		StartTime: params.StartTime,
		EndTime:   params.EndTime,
		Limit:     params.Limit,
		Offset:    params.Offset,
	}

	if params.ServiceName != "" {
		filter.ServiceName = &params.ServiceName
	}
	if params.Status != "" {
		filter.Status = &params.Status // Changed from Severity to Status
	}
	if params.Query != "" {
		filter.Query = &params.Query
	}

	// 지속 시간 필터 설정
	if params.MinDuration != nil && *params.MinDuration > 0 {
		fmin := float64(*params.MinDuration)
		filter.MinDuration = &fmin
	}
	if params.MaxDuration != nil && *params.MaxDuration > 0 {
		fmax := float64(*params.MaxDuration)
		filter.MaxDuration = &fmax
	}

	// 쿼리 실행
	result, err := c.traceService.QueryTraces(filter)
	if err != nil {
		c.logger.Error().Err(err).Msg("트레이스 쿼리 실패")
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    http.StatusInternalServerError,
				Message: "트레이스 조회 중 오류가 발생했습니다",
			},
		})
		return
	}

	// 서비스 목록 추출
	services := make(map[string]bool)
	var totalDuration float64 // float64 타입으로 변경
	for _, trace := range result.Traces {
		if trace.ServiceName != "" {
			services[trace.ServiceName] = true
		}
		totalDuration += trace.Duration
	}

	servicesList := make([]string, 0, len(services))
	for service := range services {
		servicesList = append(servicesList, service)
	}

	// 응답 구성
	response := dto.TracesResponse{
		Traces: result.Traces,
		Pagination: dto.Pagination{
			Total:  result.Total,
			Limit:  params.Limit,
			Offset: params.Offset,
		},
		TimeRange: dto.TimeRange{
			StartTime: params.StartTime,
			EndTime:   params.EndTime,
		},
		Services:      servicesList,
		TotalDuration: int64(totalDuration), // Converted float64 to int64
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data:    response,
	})
}

// GetServiceMetrics godoc
//	@Summary		서비스 지표 조회
//	@Description	서비스별 성능 지표를 조회합니다
//	@Tags			metrics
//	@Accept			json
//	@Produce		json
//	@Param			startTime	query		int		false	"시작 시간 (밀리초 타임스탬프)"
//	@Param			endTime		query		int		false	"종료 시간 (밀리초 타임스탬프)"
//	@Param			serviceName	query		string	false	"서비스 이름(선택적)"
//	@Success		200			{object}	dto.Response{data=dto.ServiceMetricsResponse}
//	@Failure		400			{object}	dto.Response
//	@Failure		500			{object}	dto.Response
//	@Router			/metrics/services [get]
func (c *TraceController) GetServiceMetrics(ctx *gin.Context) {
	// 시간 범위 파싱
	startTimeStr := ctx.DefaultQuery("startTime", "")
	endTimeStr := ctx.DefaultQuery("endTime", "")
	serviceName := ctx.Query("serviceName")

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

	// 서비스 지표를 가져오는 기능은 아직 구현되지 않았으므로 임시 응답을 반환합니다
	// TODO: 실제 서비스 지표 조회 기능 구현

	// 임시 응답 데이터
	metrics := []dto.ServiceMetric{
		{
			Name:         "api-service",
			RequestCount: 1250,
			ErrorCount:   45,
			AvgLatency:   125.5,
			P95Latency:   350.2,
			P99Latency:   520.8,
			ErrorRate:    3.6,
		},
		{
			Name:         "auth-service",
			RequestCount: 860,
			ErrorCount:   12,
			AvgLatency:   75.3,
			P95Latency:   180.1,
			P99Latency:   250.4,
			ErrorRate:    1.4,
		},
	}
	if serviceName != "" {
		filteredMetrics := []dto.ServiceMetric{}
		for _, metric := range metrics {
			if metric.Name == serviceName {
				filteredMetrics = append(filteredMetrics, metric)
			}
		}
		metrics = filteredMetrics
	}

	// 총 요청 및 오류 계산
	var totalRequests, totalErrors int64
	var totalLatency float64
	for _, metric := range metrics {
		totalRequests += metric.RequestCount
		totalErrors += metric.ErrorCount
		totalLatency += float64(metric.RequestCount) * metric.AvgLatency
	}

	// 평균 지연 시간 계산
	var avgLatency float64
	if totalRequests > 0 {
		avgLatency = totalLatency / float64(totalRequests)
	}

	// 오류율 계산
	var errorPercentage float64
	if totalRequests > 0 {
		errorPercentage = float64(totalErrors) / float64(totalRequests) * 100
	}

	response := dto.ServiceMetricsResponse{
		Services: metrics,
		TimeRange: dto.TimeRange{
			StartTime: startTime,
			EndTime:   endTime,
		},
		TotalRequests:   totalRequests,
		TotalErrors:     totalErrors,
		AvgLatency:      avgLatency,
		ErrorPercentage: errorPercentage,
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data:    response,
	})
}
