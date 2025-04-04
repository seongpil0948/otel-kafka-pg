package processor

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	protobuf "github.com/golang/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/seongpil0948/otel-kafka-pg/modules/common/logger"
	logDomain "github.com/seongpil0948/otel-kafka-pg/modules/log/domain"
	traceDomain "github.com/seongpil0948/otel-kafka-pg/modules/trace/domain"

	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	logspb "go.opentelemetry.io/proto/otlp/logs/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
)

// Processor는 메시지 처리를 위한 인터페이스입니다.
type Processor interface {
	// DecompressMessage는 압축된 메시지를 압축 해제합니다.
	DecompressMessage(data []byte) ([]byte, error)
	
	// ProcessTraceData는 트레이스 데이터를 처리합니다.
	ProcessTraceData(data []byte) ([]traceDomain.TraceItem, error)
	
	// ProcessLogData는 로그 데이터를 처리합니다.
	ProcessLogData(data []byte) ([]logDomain.LogItem, error)
}

// ProtoProcessor는 프로토콜 버퍼 형식의 메시지를 처리하는 구현체입니다.
type ProtoProcessor struct {
	log logger.Logger
}

// NewProcessor는 새 프로토콜 버퍼 프로세서 인스턴스를 생성합니다.
func NewProcessor() Processor {
	return &ProtoProcessor{
		log: logger.GetLogger(),
	}
}

// DecompressMessage는 메시지를 압축 해제합니다.
func (p *ProtoProcessor) DecompressMessage(data []byte) ([]byte, error) {
	// Snappy로 압축된 메시지인지 확인
	if len(data) > 4 && data[0] == 0xff && data[1] == 0x06 && data[2] == 0x00 && data[3] == 0x00 {
		return snappy.Decode(nil, data)
	}
	
	// 압축되지 않은 메시지
	return data, nil
}

// ProcessTraceData는 Protocol Buffer 형식의 트레이스 데이터를 처리합니다.
func (p *ProtoProcessor) ProcessTraceData(data []byte) ([]traceDomain.TraceItem, error) {
	traces := []traceDomain.TraceItem{}
	
	// OTLP ExportTraceServiceRequest 디코딩 시도
	requestData := &coltracepb.ExportTraceServiceRequest{}
	if err := protobuf.Unmarshal(data, requestData); err != nil {
		// 일반 TracesData 형식 시도
		tracesData := &tracepb.TracesData{}
		if err := protobuf.Unmarshal(data, tracesData); err != nil {
			p.log.Error().Err(err).Msg("트레이스 데이터 디코딩 실패: 지원되지 않는 형식")
			return traces, err
		}
		
		// ProcessResourceSpans 호출하여 직접 변환
		for _, resourceSpans := range tracesData.ResourceSpans {
			convertedTraces := p.ProcessResourceSpans(resourceSpans)
			traces = append(traces, convertedTraces...)
		}
		return traces, nil
	}
	
	// ExportTraceServiceRequest의 ResourceSpans 처리
	for _, resourceSpans := range requestData.ResourceSpans {
		convertedTraces := p.ProcessResourceSpans(resourceSpans)
		traces = append(traces, convertedTraces...)
	}
	
	return traces, nil
}

// ProcessResourceSpans는 ResourceSpans를 TraceItem으로 변환합니다.
func (p *ProtoProcessor) ProcessResourceSpans(resourceSpans *tracepb.ResourceSpans) []traceDomain.TraceItem {
	traces := []traceDomain.TraceItem{}
	
	// 리소스 속성 추출
	resourceAttributes := make(map[string]interface{})
	serviceName := "unknown"
	
	if resourceSpans.Resource != nil {
		for _, attr := range resourceSpans.Resource.Attributes {
			if attr.Key == "service.name" && attr.Value != nil && attr.Value.GetStringValue() != "" {
				serviceName = attr.Value.GetStringValue()
			}
			resourceAttributes[attr.Key] = p.getAttributeValue(attr.Value)
		}
	}
	
	// ScopeSpans 처리
	for _, scopeSpans := range resourceSpans.ScopeSpans {
		// Span 처리
		for _, span := range scopeSpans.Spans {
			attributes := make(map[string]interface{})
			
			// Span 속성 추출
			for _, attr := range span.Attributes {
				attributes[attr.Key] = p.getAttributeValue(attr.Value)
			}
			
			// 리소스 속성 병합
			for k, v := range resourceAttributes {
				attributes[k] = v
			}
			
			// 상태 코드 변환
			status := "UNSET"
			if span.Status != nil {
				switch span.Status.Code {
				case tracepb.Status_STATUS_CODE_OK:
					status = "OK"
				case tracepb.Status_STATUS_CODE_ERROR:
					status = "ERROR"
				}
			}
			
			// TraceItem 생성
			traceItem := traceDomain.TraceItem{
				ID:          fmt.Sprintf("%s-%s", hex.EncodeToString(span.TraceId), hex.EncodeToString(span.SpanId)),
				TraceID:     hex.EncodeToString(span.TraceId),
				SpanID:      hex.EncodeToString(span.SpanId),
				ParentSpanID: hex.EncodeToString(span.ParentSpanId),
				Name:        span.Name,
				ServiceName: serviceName,
				StartTime:   int64(span.StartTimeUnixNano / 1000000), // nano → milli
				EndTime:     int64(span.EndTimeUnixNano / 1000000),   // nano → milli
				Duration:    float64(span.EndTimeUnixNano-span.StartTimeUnixNano) / 1000000,
				Status:      status,
				Attributes:  attributes,
			}
			
			traces = append(traces, traceItem)
		}
	}
	
	return traces
}

// ProcessLogData는 Protocol Buffer 형식의 로그 데이터를 처리합니다.
func (p *ProtoProcessor) ProcessLogData(data []byte) ([]logDomain.LogItem, error) {
	logs := []logDomain.LogItem{}
	
	// OTLP ExportLogsServiceRequest 디코딩 시도
	requestData := &collogspb.ExportLogsServiceRequest{}
	if err := protobuf.Unmarshal(data, requestData); err != nil {
		// 일반 LogsData 형식 시도
		logsData := &logspb.LogsData{}
		if err := protobuf.Unmarshal(data, logsData); err != nil {
			p.log.Error().Err(err).Msg("로그 데이터 디코딩 실패: 지원되지 않는 형식")
			return logs, err
		}
		
		// ProcessResourceLogs 호출하여 직접 변환
		for _, resourceLogs := range logsData.ResourceLogs {
			convertedLogs := p.ProcessResourceLogs(resourceLogs)
			logs = append(logs, convertedLogs...)
		}
		return logs, nil
	}
	
	// ExportLogsServiceRequest의 ResourceLogs 처리
	for _, resourceLogs := range requestData.ResourceLogs {
		convertedLogs := p.ProcessResourceLogs(resourceLogs)
		logs = append(logs, convertedLogs...)
	}
	
	return logs, nil
}

// ProcessResourceLogs는 ResourceLogs를 LogItem으로 변환합니다.
func (p *ProtoProcessor) ProcessResourceLogs(resourceLogs *logspb.ResourceLogs) []logDomain.LogItem {
	logs := []logDomain.LogItem{}
	
	// 리소스 속성 추출
	resourceAttributes := make(map[string]interface{})
	serviceName := "unknown"
	
	if resourceLogs.Resource != nil {
		for _, attr := range resourceLogs.Resource.Attributes {
			if attr.Key == "service.name" && attr.Value != nil && attr.Value.GetStringValue() != "" {
				serviceName = attr.Value.GetStringValue()
			}
			resourceAttributes[attr.Key] = p.getAttributeValue(attr.Value)
		}
	}
	
	// ScopeLogs 처리
	for _, scopeLogs := range resourceLogs.ScopeLogs {
		// LogRecords 처리
		for _, logRecord := range scopeLogs.LogRecords {
			attributes := make(map[string]interface{})
			
			// 로그 속성 추출
			for _, attr := range logRecord.Attributes {
				attributes[attr.Key] = p.getAttributeValue(attr.Value)
			}
			
			// 리소스 속성 병합
			for k, v := range resourceAttributes {
				if _, exists := attributes[k]; !exists {
					attributes[k] = v
				}
			}
			
			// 스코프 속성 추가
			if scopeLogs.Scope != nil {
				attributes["scope.name"] = scopeLogs.Scope.Name
				attributes["scope.version"] = scopeLogs.Scope.Version
				
				for _, attr := range scopeLogs.Scope.Attributes {
					key := fmt.Sprintf("scope.attr.%s", attr.Key)
					attributes[key] = p.getAttributeValue(attr.Value)
				}
			}
			
			// 메시지 본문 추출
			message := ""
			if logRecord.Body != nil {
				switch bv := logRecord.Body.Value.(type) {
				case *commonpb.AnyValue_StringValue:
					message = bv.StringValue
				case *commonpb.AnyValue_KvlistValue:
					message = p.formatKVListMessage(bv.KvlistValue)
				default:
					// 바디에서 메시지를 추출할 수 없는 경우, 속성에서 메시지 찾기
					if msg, ok := attributes["message"]; ok {
						if msgStr, ok := msg.(string); ok {
							message = msgStr
						}
					}
				}
			}
			
			// 심각도 변환
			severity := "INFO"
			if logRecord.SeverityText != "" {
				severity = logRecord.SeverityText
			} else {
				severity = logDomain.SeverityNumberToText(int(logRecord.SeverityNumber))
			}
			
			// LogItem 생성
			id := p.generateLogID(logRecord.TimeUnixNano, logRecord.TraceId, logRecord.SpanId)
			logItem := logDomain.LogItem{
				ID:          id,
				Timestamp:   int64(logRecord.TimeUnixNano / 1000000), // nano → milli
				ServiceName: serviceName,
				Message:     message,
				Severity:    severity,
				TraceID:     hex.EncodeToString(logRecord.TraceId),
				SpanID:      hex.EncodeToString(logRecord.SpanId),
				Attributes:  attributes,
			}
			
			logs = append(logs, logItem)
		}
	}
	
	return logs
}

// getAttributeValue는 속성 값을 적절한 Go 타입으로 변환합니다.
func (p *ProtoProcessor) getAttributeValue(value *commonpb.AnyValue) interface{} {
	if value == nil {
		return nil
	}
	
	switch v := value.Value.(type) {
	case *commonpb.AnyValue_StringValue:
		return v.StringValue
	case *commonpb.AnyValue_BoolValue:
		return v.BoolValue
	case *commonpb.AnyValue_IntValue:
		return v.IntValue
	case *commonpb.AnyValue_DoubleValue:
		return v.DoubleValue
	case *commonpb.AnyValue_ArrayValue:
		if v.ArrayValue == nil || len(v.ArrayValue.Values) == 0 {
			return []interface{}{}
		}
		
		values := make([]interface{}, len(v.ArrayValue.Values))
		for i, val := range v.ArrayValue.Values {
			values[i] = p.getAttributeValue(val)
		}
		return values
	case *commonpb.AnyValue_KvlistValue:
		if v.KvlistValue == nil || len(v.KvlistValue.Values) == 0 {
			return map[string]interface{}{}
		}
		
		kvMap := make(map[string]interface{})
		for _, kv := range v.KvlistValue.Values {
			kvMap[kv.Key] = p.getAttributeValue(kv.Value)
		}
		return kvMap
	case *commonpb.AnyValue_BytesValue:
		return hex.EncodeToString(v.BytesValue)
	default:
		return nil
	}
}

// formatKVListMessage는 KeyValue 리스트 형태의 로그 바디를 문자열로 변환합니다.
func (p *ProtoProcessor) formatKVListMessage(kvlist *commonpb.KeyValueList) string {
	if kvlist == nil || len(kvlist.Values) == 0 {
		return ""
	}
	
	// 메시지 또는 이벤트 관련 필드 찾기
	for _, kv := range kvlist.Values {
		if kv.Key == "message" || kv.Key == "msg" || kv.Key == "event" || kv.Key == "log" {
			if val := p.getAttributeValue(kv.Value); val != nil {
				if str, ok := val.(string); ok && str != "" {
					return str
				}
			}
		}
	}
	
	// 메시지 필드가 없으면 첫 번째 필드 사용
	if len(kvlist.Values) > 0 {
		kv := kvlist.Values[0]
		if val := p.getAttributeValue(kv.Value); val != nil {
			return fmt.Sprintf("%s: %v", kv.Key, val)
		}
	}
	
	return "{키-값 목록 로그}"
}

// generateLogID는 로그 ID를 생성합니다.
func (p *ProtoProcessor) generateLogID(timeNano uint64, traceID, spanID []byte) string {
	// 타임스탬프(밀리초) + 트레이스 ID + 스팬 ID + 8자리 랜덤 문자열
	timeMs := timeNano / 1000000
	
	// 타임스탬프 + 트레이스/스팬 ID 해시
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%d", timeMs)))
	if len(traceID) > 0 {
		h.Write(traceID)
	}
	if len(spanID) > 0 {
		h.Write(spanID)
	}
	// 추가 엔트로피를 위한 현재 나노초
	h.Write([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	
	hashBytes := h.Sum(nil)
	hashStr := hex.EncodeToString(hashBytes[:8]) // 16자리 해시
	
	return fmt.Sprintf("%d-%s", timeMs, hashStr)
}