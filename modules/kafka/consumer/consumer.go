package consumer

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/seongpil0948/otel-kafka-pg/modules/common/config"
	"github.com/seongpil0948/otel-kafka-pg/modules/common/logger"
	"github.com/seongpil0948/otel-kafka-pg/modules/kafka/processor"
	logDomain "github.com/seongpil0948/otel-kafka-pg/modules/log/domain"
	logService "github.com/seongpil0948/otel-kafka-pg/modules/log/service"
	traceDomain "github.com/seongpil0948/otel-kafka-pg/modules/trace/domain"
	traceService "github.com/seongpil0948/otel-kafka-pg/modules/trace/service"
)

// MessageBuffer는 메시지 버퍼 구조체입니다.
type MessageBuffer struct {
	Traces       []traceDomain.TraceItem
	Logs         []logDomain.LogItem
	LastFlushTime time.Time
	mu           sync.Mutex
}

// Consumer는 Kafka 소비자 인터페이스입니다.
type Consumer interface {
	Start(ctx context.Context) error
	Stop() error
	FlushBuffer() error
}

// KafkaConsumer는 Kafka 소비자 구현체입니다.
type KafkaConsumer struct {
	client        *kafka.Consumer
	processor     processor.Processor
	traceService  traceService.TraceService
	logService    logService.LogService
	cfg           *config.Config
	log           logger.Logger
	messageBuffer MessageBuffer
	flushTicker   *time.Ticker
	isRunning     bool
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewConsumer는 새 Kafka 소비자 인스턴스를 생성합니다.
func NewConsumer(
	proc processor.Processor, 
	traceService traceService.TraceService,
	logService logService.LogService,
) Consumer {
	cfg := config.GetConfig()
	log := logger.GetLogger()

	return &KafkaConsumer{
		processor:     proc,
		traceService:  traceService,
		logService:    logService,
		cfg:           cfg,
		log:           log,
		messageBuffer: MessageBuffer{
			Traces:       []traceDomain.TraceItem{},
			Logs:         []logDomain.LogItem{},
			LastFlushTime: time.Now(),
		},
		isRunning: false,
	}
}

// Start는 Kafka 소비자를 시작합니다.
func (c *KafkaConsumer) Start(ctx context.Context) error {
	if c.isRunning {
		c.log.Info().Msg("Kafka consumer is already running")
		return nil
	}

	c.ctx, c.cancel = context.WithCancel(ctx)

	// Kafka 설정
	kafkaConfig := &kafka.ConfigMap{
		"bootstrap.servers":       c.cfg.Kafka.Brokers,
		"group.id":                c.cfg.Kafka.GroupID,
		"client.id":               c.cfg.Kafka.ClientID,
		"auto.offset.reset":       "latest",
		"session.timeout.ms":      30000,
		"heartbeat.interval.ms":   5000,
		"enable.auto.commit":      true,
		"auto.commit.interval.ms": 5000,
	}

	// 소비자 생성
	consumer, err := kafka.NewConsumer(kafkaConfig)
	if err != nil {
		c.log.Error().Err(err).Msg("Failed to create Kafka consumer")
		return err
	}
	c.client = consumer

	// 토픽 구독
	topics := []string{c.cfg.Kafka.TracesTopic, c.cfg.Kafka.LogsTopic}
	err = consumer.SubscribeTopics(topics, nil)
	if err != nil {
		c.log.Error().Err(err).Msg("Failed to subscribe to Kafka topics")
		return err
	}
	c.log.Info().Strs("topics", topics).Msg("Subscribed to Kafka topics")

	// 플러시 타이머 시작
	c.flushTicker = time.NewTicker(time.Duration(c.cfg.Kafka.FlushInterval) * time.Millisecond)

	// 메시지 수신 고루틴 시작
	go c.consumeMessages()

	c.isRunning = true
	c.log.Info().Msg("Kafka consumer started successfully")

	return nil
}

// 메시지 소비 함수
func (c *KafkaConsumer) consumeMessages() {
	for {
		select {
		case <-c.ctx.Done():
			c.log.Info().Msg("Stopping Kafka consumer")
			return

		case <-c.flushTicker.C:
			// 주기적으로 버퍼 플러시
			c.FlushBuffer()

		default:
			// 메시지 폴링
			msg := c.client.Poll(100) // Poll은 단일 값만 반환함

			if msg == nil {
				continue
			}

			switch ev := msg.(type) {
			case *kafka.Message:
				// 메시지 처리
				msg := c.client.Poll(100)
				if msg == nil {
						continue
				}

				// 버퍼 사이즈가 임계값을 초과하면 플러시
				c.messageBuffer.mu.Lock()
				tracesLen := len(c.messageBuffer.Traces)
				logsLen := len(c.messageBuffer.Logs)
				timeSinceLastFlush := time.Since(c.messageBuffer.LastFlushTime)
				c.messageBuffer.mu.Unlock()

				if tracesLen >= c.cfg.Kafka.BatchSize || logsLen >= c.cfg.Kafka.BatchSize || 
					timeSinceLastFlush.Milliseconds() >= int64(c.cfg.Kafka.FlushInterval) {
					c.FlushBuffer()
				}

			case kafka.Error:
				// Kafka 에러 처리
				c.log.Error().
					Str("code", ev.Code().String()).
					Msg(ev.Error())

				// 치명적인 에러인 경우 재연결
				if ev.Code() == kafka.ErrAllBrokersDown ||
					ev.Code() == kafka.ErrNetworkException {
					c.log.Error().Msg("Critical Kafka error, attempting to reconnect")
					c.reconnect()
					return
				}
			}
		}
	}
}

// 메시지 처리 함수
func (c *KafkaConsumer) processMessage(msg *kafka.Message) error {
	if msg.Value == nil {
		return nil
	}

	// 메시지 압축 해제
	decompressedValue, err := c.processor.DecompressMessage(msg.Value)
	if err != nil {
		return err
	}

	// 토픽에 따른 메시지 처리
	topic := *msg.TopicPartition.Topic

	if topic == c.cfg.Kafka.TracesTopic {
		// Protobuf로 디코딩 (여기서는 JSON으로 단순화)
		var data map[string]interface{}
		if err := json.Unmarshal(decompressedValue, &data); err != nil {
			return err
		}

		traces, err := c.processor.ProcessTraceData(data)
		if err != nil {
			return err
		}
		
		if len(traces) > 0 {
			c.messageBuffer.mu.Lock()
			c.messageBuffer.Traces = append(c.messageBuffer.Traces, traces...)
			c.messageBuffer.mu.Unlock()
			c.log.Debug().Int("count", len(traces)).Msg("Processed trace data")
		}
	} else if topic == c.cfg.Kafka.LogsTopic {
		// Protobuf로 디코딩 (여기서는 JSON으로 단순화)
		var data map[string]interface{}
		if err := json.Unmarshal(decompressedValue, &data); err != nil {
			return err
		}

		logs, err := c.processor.ProcessLogData(data)
		if err != nil {
			return err
		}
		
		if len(logs) > 0 {
			c.messageBuffer.mu.Lock()
			c.messageBuffer.Logs = append(c.messageBuffer.Logs, logs...)
			c.messageBuffer.mu.Unlock()
			c.log.Debug().Int("count", len(logs)).Msg("Processed log data")
		}
	}

	return nil
}

// FlushBuffer는 버퍼에 있는 메시지를 데이터베이스에 저장합니다.
func (c *KafkaConsumer) FlushBuffer() error {
	c.messageBuffer.mu.Lock()
	traces := c.messageBuffer.Traces
	logs := c.messageBuffer.Logs
	c.messageBuffer.Traces = []traceDomain.TraceItem{}
	c.messageBuffer.Logs = []logDomain.LogItem{}
	c.messageBuffer.LastFlushTime = time.Now()
	c.messageBuffer.mu.Unlock()

	// 트레이스 데이터 저장
	if len(traces) > 0 {
		c.log.Info().Int("count", len(traces)).Msg("Flushing trace data to database")
		err := c.traceService.SaveTraces(traces)
		if err != nil {
			c.log.Error().Err(err).Msg("Error saving traces")
			// 실패 시 다시 버퍼에 추가
			c.messageBuffer.mu.Lock()
			c.messageBuffer.Traces = append(c.messageBuffer.Traces, traces...)
			c.messageBuffer.mu.Unlock()
		}
	}

	// 로그 데이터 저장
	if len(logs) > 0 {
		c.log.Info().Int("count", len(logs)).Msg("Flushing log data to database")
		err := c.logService.SaveLogs(logs)
		if err != nil {
			c.log.Error().Err(err).Msg("Error saving logs")
			// 실패 시 다시 버퍼에 추가
			c.messageBuffer.mu.Lock()
			c.messageBuffer.Logs = append(c.messageBuffer.Logs, logs...)
			c.messageBuffer.mu.Unlock()
		}
	}

	return nil
}

// Stop은 Kafka 소비자를 중지합니다.
func (c *KafkaConsumer) Stop() error {
	if !c.isRunning || c.client == nil {
		c.log.Info().Msg("Kafka consumer is not running")
		return nil
	}

	// 컨텍스트 취소
	if c.cancel != nil {
		c.cancel()
	}

	// 마지막으로 버퍼 플러시
	err := c.FlushBuffer()
	if err != nil {
		c.log.Error().Err(err).Msg("Error flushing buffer during shutdown")
	}

	// 타이머 정리
	if c.flushTicker != nil {
		c.flushTicker.Stop()
	}

	// 컨슈머 연결 해제
	err = c.client.Close()
	if err != nil {
		c.log.Error().Err(err).Msg("Error closing Kafka consumer")
		return err
	}

	c.client = nil
	c.isRunning = false
	c.log.Info().Msg("Kafka consumer stopped successfully")
	return nil
}

// 재연결 함수
func (c *KafkaConsumer) reconnect() {
	c.Stop()
	time.Sleep(5 * time.Second) // 재연결 전 잠시 대기
	
	ctx := context.Background()
	if c.ctx != nil {
		ctx = c.ctx
	}
	
	err := c.Start(ctx)
	if err != nil {
		c.log.Error().Err(err).Msg("Failed to restart Kafka consumer")
	}
}