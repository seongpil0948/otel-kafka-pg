package processor

import (
	"encoding/json"
	"fmt"

	"github.com/golang/snappy"
	"github.com/seongpil0948/otel-kafka-pg/modules/common/logger"
	"github.com/seongpil0948/otel-kafka-pg/modules/log/domain"
	traceModel "github.com/seongpil0948/otel-kafka-pg/modules/trace/domain"
)

// Processor는 메시지 처리를 위한 인터페이스입니다.
type Processor interface {
	// DecompressMessage는 메시지를 압축 해제합니다.
	DecompressMessage(data []byte) ([]byte, error)
	
	// ProcessTraceData는 트레이스 데이터를 처리합니다.
	ProcessTraceData(data map[string]interface{}) ([]traceModel.TraceItem, error)
	
	// ProcessLogData는 로그 데이터를 처리합니다.
	ProcessLogData(data map[string]interface{}) ([]domain.LogItem, error)
}

// KafkaProcessor는 Kafka 메시지 처리기 구현체입니다.
type KafkaProcessor struct {
	log logger.Logger
}

// NewProcessor는 새 프로세서 인스턴스를 생성합니다.
func NewProcessor() Processor {
	return &KafkaProcessor{
		log: logger.GetLogger(),
	}
}

// DecompressMessage는 메시지를 압축 해제합니다.
func (p *KafkaProcessor) DecompressMessage(data []byte) ([]byte, error) {
	// Snappy로 압축된 메시지인지 확인
	if len(data) > 4 && data[0] == 0xff && data[1] == 0x06 && data[2] == 0x00 && data[3] == 0x00 {
		return snappy.Decode(nil, data)
	}
	
	// 압축되지 않은 메시지
	return data, nil
}

// MapAttributes는 속성 배열을 키-값 맵으로 변환합니다.
func (p *KafkaProcessor) MapAttributes(attributes []interface{}) map[string]interface{} {
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
			if v, exists := valueObj["stringValue"]; exists {
				value = v
			} else if v, exists := valueObj["boolValue"]; exists {
				value = v
			} else if v, exists := valueObj["intValue"]; exists {
				value = v
			} else if v, exists := valueObj["doubleValue"]; exists {
				value = v
			} else if v, exists := valueObj["arrayValue"]; exists {
				value = v
			} else if v, exists := valueObj["kvlistValue"]; exists {
				value = v
			}
		}

		result[key] = value
	}

	return result
}

// ProcessTraceData는 트레이스 데이터를 처리합니다.
func (p *KafkaProcessor) ProcessTraceData(data map[string]interface{}) ([]traceModel.TraceItem, error) {
	traces := []traceModel.TraceItem{}

	resourceSpans, ok := data["resourceSpans"].([]interface{})
	if !ok {
		p.log.Debug().Msg("No resourceSpans found")
		return traces, nil
	}

	for _, resourceSpan := range resourceSpans {
		rs, ok := resourceSpan.(map[string]interface{})
		if !ok {
			continue
		}

		resourceObj, ok := rs["resource"].(map[string]interface{})
		if !ok {
			resourceObj = map[string]interface{}{}
		}

		resourceAttrs, ok := resourceObj["attributes"].([]interface{})
		if !ok {
			resourceAttrs = []interface{}{}
		}

		resourceAttributes := p.MapAttributes(resourceAttrs)
		serviceName := "unknown"
		if sn, ok := resourceAttributes["service.name"].(string); ok {
			serviceName = sn
		}

		scopeSpans, ok := rs["scopeSpans"].([]interface{})
		if !ok {
			continue
		}

		for _, scopeSpan := range scopeSpans {
			ss, ok := scopeSpan.(map[string]interface{})
			if !ok {
				continue
			}

			scope, ok := ss["scope"].(map[string]interface{})
			if !ok {
				scope = map[string]interface{}{}
			}

			spans, ok := ss["spans"].([]interface{})
			if !ok {
				continue
			}

			for _, span := range spans {
				s, ok := span.(map[string]interface{})
				if !ok {
					continue
				}

				// 필수 필드 가져오기
				traceID, _ := s["traceId"].(string)
				spanID, _ := s["spanId"].(string)
				parentSpanID, _ := s["parentSpanId"].(string)
				name, _ := s["name"].(string)
				kind, _ := s["kind"].(float64)

				// 타임스탬프는 문자열 또는 숫자일 수 있음
				var startTimeUnixNano, endTimeUnixNano int64
				if st, ok := s["startTimeUnixNano"].(float64); ok {
					startTimeUnixNano = int64(st)
				} else if st, ok := s["startTimeUnixNano"].(string); ok {
					json.Unmarshal([]byte(st), &startTimeUnixNano)
				}

				if et, ok := s["endTimeUnixNano"].(float64); ok {
					endTimeUnixNano = int64(et)
				} else if et, ok := s["endTimeUnixNano"].(string); ok {
					json.Unmarshal([]byte(et), &endTimeUnixNano)
				}

				// 밀리초로 변환
				startTime := startTimeUnixNano / 1000000
				endTime := endTimeUnixNano / 1000000
				duration := float64(endTimeUnixNano-startTimeUnixNano) / 1000000

				status := "UNSET"
				if statusObj, ok := s["status"].(map[string]interface{}); ok {
					if code, ok := statusObj["code"].(string); ok {
						status = code
					}
				}

				// 속성 가져오기
				spanAttrs, ok := s["attributes"].([]interface{})
				if !ok {
					spanAttrs = []interface{}{}
				}

				attributes := p.MapAttributes(spanAttrs)
				// 리소스 속성 병합
				for k, v := range resourceAttributes {
					attributes[k] = v
				}

				// TraceItem 객체 생성
				traceItem := traceModel.TraceItem{
					ID:          traceID + "-" + spanID,
					TraceID:     traceID,
					SpanID:      spanID,
					ParentSpanID: parentSpanID,
					Name:        name,
					ServiceName: serviceName,
					StartTime:   startTime,
					EndTime:     endTime,
					Duration:    duration,
					Status:      status,
					Attributes:  attributes,
				}

				traces = append(traces, traceItem)
			}
		}
	}

	return traces, nil
}

// ProcessLogData는 로그 데이터를 처리합니다.
func (p *KafkaProcessor) ProcessLogData(data map[string]interface{}) ([]domain.LogItem, error) {
	logs := []domain.LogItem{}

	resourceLogs, ok := data["resourceLogs"].([]interface{})
	if !ok {
		p.log.Debug().Msg("No resourceLogs found")
		return logs, nil
	}

	for _, resourceLog := range resourceLogs {
		rl, ok := resourceLog.(map[string]interface{})
		if !ok {
			continue
		}

		resourceObj, ok := rl["resource"].(map[string]interface{})
		if !ok {
			resourceObj = map[string]interface{}{}
		}

		resourceAttrs, ok := resourceObj["attributes"].([]interface{})
		if !ok {
			resourceAttrs = []interface{}{}
		}

		resourceAttributes := p.MapAttributes(resourceAttrs)
		serviceName := "unknown"
		if sn, ok := resourceAttributes["service.name"].(string); ok {
			serviceName = sn
		}

		scopeLogs, ok := rl["scopeLogs"].([]interface{})
		if !ok {
			continue
		}

		for _, scopeLog := range scopeLogs {
			sl, ok := scopeLog.(map[string]interface{})
			if !ok {
				continue
			}

			scope, ok := sl["scope"].(map[string]interface{})
			if !ok {
				scope = map[string]interface{}{}
			}

			logRecords, ok := sl["logRecords"].([]interface{})
			if !ok {
				continue
			}

			for _, logRecord := range logRecords {
				lr, ok := logRecord.(map[string]interface{})
				if !ok {
					continue
				}

				// 타임스탬프
				var timeUnixNano int64
				if t, ok := lr["timeUnixNano"].(float64); ok {
					timeUnixNano = int64(t)
				} else if t, ok := lr["timeUnixNano"].(string); ok {
					json.Unmarshal([]byte(t), &timeUnixNano)
				}
				timestamp := timeUnixNano / 1000000 // 밀리초로 변환

				// 메시지 본문
				body := ""
				if bodyObj, ok := lr["body"].(map[string]interface{}); ok {
					if sv, ok := bodyObj["stringValue"].(string); ok {
						body = sv
					} else {
						bodyBytes, _ := json.Marshal(bodyObj)
						body = string(bodyBytes)
					}
				}

				// 심각도
				severity := "INFO"
				if severityText, ok := lr["severityText"].(string); ok && severityText != "" {
					severity = severityText
				} else if severityNumber, ok := lr["severityNumber"].(float64); ok {
					severity = domain.SeverityNumberToText(int(severityNumber))
				}

				// 트레이스 및 스팬 ID
				traceID := ""
				spanID := ""
				if tid, ok := lr["traceId"].(string); ok {
					traceID = tid
				}
				if sid, ok := lr["spanId"].(string); ok {
					spanID = sid
				}

				// 속성
				logAttrs, ok := lr["attributes"].([]interface{})
				if !ok {
					logAttrs = []interface{}{}
				}
				attributes := p.MapAttributes(logAttrs)

				// 속성에서 트레이스/스팬 ID 확인
				if traceID == "" {
					if tid, ok := attributes["trace_id"].(string); ok {
						traceID = tid
					}
				}
				if spanID == "" {
					if sid, ok := attributes["span_id"].(string); ok {
						spanID = sid
					}
				}

				// 리소스 속성 병합
				for k, v := range resourceAttributes {
					attributes[k] = v
				}

				// LogItem 객체 생성
				id := fmt.Sprintf("%d-%s", timestamp, generateID(8))
				logItem := domain.LogItem{
					ID:          id,
					Timestamp:   timestamp,
					ServiceName: serviceName,
					Message:     body,
					Severity:    severity,
					TraceID:     traceID,
					SpanID:      spanID,
					Attributes:  attributes,
				}

				logs = append(logs, logItem)
			}
		}
	}

	return logs, nil
}

// generateID는 지정된 길이의 랜덤 ID를 생성합니다.
func generateID(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[int(fmt.Sprintf("%d", i)[0])%len(charset)]
	}
	return string(b)
}