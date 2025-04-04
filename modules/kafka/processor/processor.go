package processor

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/seongpil0948/otel-kafka-pg/modules/common/logger"
	logDomain "github.com/seongpil0948/otel-kafka-pg/modules/log/domain"
	traceDomain "github.com/seongpil0948/otel-kafka-pg/modules/trace/domain"
	logspb "github.com/seongpil0948/otel-kafka-pg/proto/gen/opentelemetry/proto/logs/v1"
	tracepb "github.com/seongpil0948/otel-kafka-pg/proto/gen/opentelemetry/proto/trace/v1"
)

// ProtoProcessor는 프로토콜 버퍼 형식의 메시지를 처리하는 구현체입니다.
type ProtoProcessor struct {
	log logger.Logger
}

// NewProtoProcessor는 새 프로토콜 버퍼 프로세서 인스턴스를 생성합니다.
func NewProtoProcessor() Processor {
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

// MapAttributes는 속성 배열을 키-값 맵으로 변환합니다.
// 기존 코드와의 호환성을 위해 유지
func (p *ProtoProcessor) MapAttributes(attributes []interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for _, attr := range attributes {
		attrMap, ok := attr.(map[string]interface{})
		if !ok || attrMap["key"] == nil {
			continue
		}

		key, ok := attrMap["key"].(string)
		if !ok {
			continue
		}

		var value interface{} = nil
		valueObj, hasValue := attrMap["value"].(map[string]interface{})
		
		if hasValue {
			// 타입 스위치를 사용하여 더 명확하게 처리
            for valueType, _ := range map[string]string{
                "stringValue": "string", 
                "boolValue": "bool",
                "intValue": "int",
                "doubleValue": "double",
                "arrayValue": "array",
                "kvlistValue": "kvlist",
            } {
                if v, exists := valueObj[valueType]; exists {
                    value = v
                    break
                }
            }
		}

		result[key] = value
	}

	return result
}

// ProcessTraceData는 Protocol Buffer 형식의 트레이스 데이터를 처리합니다.
func (p *ProtoProcessor) ProcessTraceData(data []byte) ([]traceDomain.TraceItem, error) {
	traces := []traceDomain.TraceItem{}
	
	// JSON 형식인 경우 (기존 방식과의 호환성을 위해)
	if len(data) > 0 && data[0] == '{' {
		var jsonData map[string]interface{}
		// 기존 JSON 처리 로직을 호출 (필요시 구현)
		return p.processJSONTraceData(jsonData)
	}
	
	// Protocol Buffer 디코딩
	tracesData := &tracepb.TracesData{}
	if err := proto.Unmarshal(data, tracesData); err != nil {
		p.log.Error().Err(err).Msg("트레이스 데이터 디코딩 실패")
		return traces, err
	}
	
	// ResourceSpans 처리
	for _, resourceSpans := range tracesData.ResourceSpans {
		// 리소스 속성 추출
		resourceAttributes := make(map[string]interface{})
		serviceName := "unknown"
		
		if resourceSpans.Resource != nil {
			for _, attr := range resourceSpans.Resource.Attributes {
				if attr.Key == "service.name" && attr.Value != nil && attr.Value.StringValue != "" {
					serviceName = attr.Value.StringValue
				}
				resourceAttributes[attr.Key] = getAttributeValue(attr.Value)
			}
		}
		
		// ScopeSpans 처리
		for _, scopeSpans := range resourceSpans.ScopeSpans {
			// Spans 처리
			for _, span := range scopeSpans.Spans {
				attributes := make(map[string]interface{})
				
				// Span 속성 추출
				for _, attr := range span.Attributes {
					attributes[attr.Key] = getAttributeValue(attr.Value)
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
					ID:          string(span.TraceId) + "-" + string(span.SpanId),
					TraceID:     string(span.TraceId),
					SpanID:      string(span.SpanId),
					ParentSpanID: string(span.ParentSpanId),
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
	}
	
	return traces, nil
}

// ProcessLogData는 Protocol Buffer 형식의 로그 데이터를 처리합니다.
func (p *ProtoProcessor) ProcessLogData(data []byte) ([]logDomain.LogItem, error) {
	logs := []logDomain.LogItem{}
	
	// JSON 형식인 경우 (기존 방식과의 호환성을 위해)
	if len(data) > 0 && data[0] == '{' {
		var jsonData map[string]interface{}
		// 기존 JSON 처리 로직을 호출 (필요시 구현)
		return p.processJSONLogData(jsonData)
	}
	
	// Protocol Buffer 디코딩
	logsData := &logspb.LogsData{}
	if err := proto.Unmarshal(data, logsData); err != nil {
		p.log.Error().Err(err).Msg("로그 데이터 디코딩 실패")
		return logs, err
	}
	
	// ResourceLogs 처리
	for _, resourceLogs := range logsData.ResourceLogs {
		// 리소스 속성 추출
		resourceAttributes := make(map[string]interface{})
		serviceName := "unknown"
		
		if resourceLogs.Resource != nil {
			for _, attr := range resourceLogs.Resource.Attributes {
				if attr.Key == "service.name" && attr.Value != nil && attr.Value.StringValue != "" {
					serviceName = attr.Value.StringValue
				}
				resourceAttributes[attr.Key] = getAttributeValue(attr.Value)
			}
		}
		
		// ScopeLogs 처리
		for _, scopeLogs := range resourceLogs.ScopeLogs {
			// LogRecords 처리
			for _, logRecord := range scopeLogs.LogRecords {
				attributes := make(map[string]interface{})
				
				// 로그 속성 추출
				for _, attr := range logRecord.Attributes {
					attributes[attr.Key] = getAttributeValue(attr.Value)
				}
				
				// 리소스 속성 병합
				for k, v := range resourceAttributes {
					attributes[k] = v
				}
				
				// 메시지 본문 추출
				message := ""
				if logRecord.Body != nil {
					message = logRecord.Body.StringValue
				}
				
				// 심각도 변환
				severity := "INFO"
				if logRecord.SeverityText != "" {
					severity = logRecord.SeverityText
				} else {
					severity = logDomain.SeverityNumberToText(int(logRecord.SeverityNumber))
				}
				
				// LogItem 생성
				id := generateID(logRecord.TimeUnixNano)
				logItem := logDomain.LogItem{
					ID:          id,
					Timestamp:   int64(logRecord.TimeUnixNano / 1000000), // nano → milli
					ServiceName: serviceName,
					Message:     message,
					Severity:    severity,
					TraceID:     string(logRecord.TraceId),
					SpanID:      string(logRecord.SpanId),
					Attributes:  attributes,
				}
				
				logs = append(logs, logItem)
			}
		}
	}
	
	return logs, nil
}

// getAttributeValue는 속성 값을 적절한 Go 타입으로 변환합니다.
func getAttributeValue(value *tracepb.AnyValue) interface{} {
	if value == nil {
		return nil
	}
	
	switch v := value.Value.(type) {
	case *tracepb.AnyValue_StringValue:
		return v.StringValue
	case *tracepb.AnyValue_BoolValue:
		return v.BoolValue
	case *tracepb.AnyValue_IntValue:
		return v.IntValue
	case *tracepb.AnyValue_DoubleValue:
		return v.DoubleValue
	case *tracepb.AnyValue_ArrayValue:
		if v.ArrayValue == nil || len(v.ArrayValue.Values) == 0 {
			return []interface{}{}
		}
		
		values := make([]interface{}, len(v.ArrayValue.Values))
		for i, val := range v.ArrayValue.Values {
			values[i] = getAttributeValue(val)
		}
		return values
	case *tracepb.AnyValue_KvlistValue:
		if v.KvlistValue == nil || len(v.KvlistValue.Values) == 0 {
			return map[string]interface{}{}
		}
		
		kvMap := make(map[string]interface{})
		for _, kv := range v.KvlistValue.Values {
			kvMap[kv.Key] = getAttributeValue(kv.Value)
		}
		return kvMap
	default:
		return nil
	}
}

// generateID는 주어진 타임스탬프를 기반으로 고유 ID를 생성합니다.
func generateID(timeNano uint64) string {
	timeMs := timeNano / 1000000
	return fmt.Sprintf("%d-%s", timeMs, generateRandomString(8))
}

// generateRandomString은 지정된 길이의 랜덤 문자열을 생성합니다.
func generateRandomString(length int) string {
	// 기존 구현과 동일
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[int(fmt.Sprintf("%d", i)[0])%len(charset)]
	}
	return string(b)
}

// 기존 JSON 처리 방식과의 호환성을 위한 메서드들
func (p *ProtoProcessor) processJSONTraceData(data map[string]interface{}) ([]traceDomain.TraceItem, error) {
	// 기존 ProcessTraceData 메서드의 로직을 여기로 이동
	// 필요시 구현
	return []traceDomain.TraceItem{}, nil
}

func (p *ProtoProcessor) processJSONLogData(data map[string]interface{}) ([]logDomain.LogItem, error) {
	// 기존 ProcessLogData 메서드의 로직을 여기로 이동
	// 필요시 구현
	return []logDomain.LogItem{}, nil
}