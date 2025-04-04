package consumer

import (
	"context"
	"fmt"
	"strings"
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
	wg            sync.WaitGroup
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

func (c *KafkaConsumer) Start(ctx context.Context) error {
	if c.isRunning {
			c.log.Info().Msg("Kafka consumer is already running")
			return nil
	}

	c.ctx, c.cancel = context.WithCancel(ctx)

	brokerList := strings.Join(c.cfg.Kafka.Brokers, ",")
	kafkaConfig := &kafka.ConfigMap{
			"bootstrap.servers":              brokerList,
			"group.id":                       c.cfg.Kafka.GroupID,
			"client.id":                      c.cfg.Kafka.ClientID,
			"auto.offset.reset":              "earliest",  // 모든 메시지 처리를 위해 earliest로 변경
			"session.timeout.ms":             30000,
			"heartbeat.interval.ms":          5000,
			"enable.auto.commit":             true,
			"auto.commit.interval.ms":        5000,
			
			// 중요: 단일 할당 전략 사용
			"partition.assignment.strategy":  "range",  // 여러 전략을 콤마로 나열하지 말고 하나만 사용
			
			// 안정성 향상을 위한 추가 설정
			"socket.keepalive.enable":        true,
			"socket.max.fails":               3,
			"reconnect.backoff.ms":           1000,
			"reconnect.backoff.max.ms":       10000,
			
			// 메타데이터 설정
			"metadata.max.age.ms":            300000,
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
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.consumeMessages()
	}()

	// 주기적으로 버퍼 플러시하는 고루틴 시작
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.periodicFlush()
	}()

	c.isRunning = true
	c.log.Info().Msg("Kafka consumer started successfully")

	return nil
}

// periodicFlush는 주기적으로 메시지 버퍼를 플러시합니다.
func (c *KafkaConsumer) periodicFlush() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-c.flushTicker.C:
			if err := c.FlushBuffer(); err != nil {
				c.log.Error().Err(err).Msg("주기적 버퍼 플러시 중 오류 발생")
			}
		}
	}
}

// 메시지 소비 함수
func (c *KafkaConsumer) consumeMessages() {
	// Kafka로부터 메시지를 수신하는 루프
	for {
		select {
		case <-c.ctx.Done():
			c.log.Info().Msg("Stopping Kafka message consumption")
			return
		default:
			// 메시지 폴링
			ev := c.client.Poll(100) // 100ms 타임아웃으로 메시지 폴링
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				// 메시지 처리
				if err := c.processMessage(e); err != nil {
					c.log.Error().Err(err).
						Str("topic", *e.TopicPartition.Topic).
						Int32("partition", e.TopicPartition.Partition).
						Str("offset", e.TopicPartition.Offset.String()).
						Msg("메시지 처리 중 오류 발생")
				}

				// 버퍼 크기 확인하여 임계값 초과 시 플러시
				c.messageBuffer.mu.Lock()
				tracesLen := len(c.messageBuffer.Traces)
				logsLen := len(c.messageBuffer.Logs)
				c.messageBuffer.mu.Unlock()

				if tracesLen >= c.cfg.Kafka.BatchSize || logsLen >= c.cfg.Kafka.BatchSize {
					if err := c.FlushBuffer(); err != nil {
						c.log.Error().Err(err).Msg("버퍼 플러시 중 오류 발생")
					}
				}

			case kafka.Error:
				// Kafka 에러 처리
				c.log.Error().
					Str("code", e.Code().String()).
					Msg("Kafka error: " + e.Error())

				// 치명적인 에러인 경우 재연결 시도
				if e.Code() == kafka.ErrAllBrokersDown ||
				   e.Code() == kafka.ErrNetworkException {
					c.log.Error().Msg("Critical Kafka error, attempting to reconnect")
					go c.reconnect()
					return
				}

			case *kafka.Stats:
				// Kafka 통계 정보 로깅
				c.log.Debug().Msg("Kafka stats received")

			default:
				// 기타 Kafka 이벤트 처리
				c.log.Debug().
					Str("event", fmt.Sprintf("%T", ev)).
					Msg("Ignored Kafka event")
			}
		}
	}
}

// 메시지 처리 함수
func (c *KafkaConsumer) processMessage(msg *kafka.Message) error {
	if msg == nil || msg.Value == nil {
		return nil
	}

	topic := *msg.TopicPartition.Topic
	
	// 메시지 압축 해제
	decompressedValue, err := c.processor.DecompressMessage(msg.Value)
	if err != nil {
		return fmt.Errorf("message decompression failed: %w", err)
	}

	// 토픽에 따른 메시지 처리
	switch topic {
	case c.cfg.Kafka.TracesTopic:
		traces, err := c.processor.ProcessTraceData(decompressedValue)
		if err != nil {
			return fmt.Errorf("trace data processing failed: %w", err)
		}
		
		if len(traces) > 0 {
			c.messageBuffer.mu.Lock()
			c.messageBuffer.Traces = append(c.messageBuffer.Traces, traces...)
			c.messageBuffer.mu.Unlock()
			c.log.Debug().Int("count", len(traces)).Msg("Processed trace data")
		}

	case c.cfg.Kafka.LogsTopic:
		logs, err := c.processor.ProcessLogData(decompressedValue)
		if err != nil {
			return fmt.Errorf("log data processing failed: %w", err)
		}
		
		if len(logs) > 0 {
			c.messageBuffer.mu.Lock()
			c.messageBuffer.Logs = append(c.messageBuffer.Logs, logs...)
			c.messageBuffer.mu.Unlock()
			c.log.Debug().Int("count", len(logs)).Msg("Processed log data")
		}

	default:
		c.log.Warn().Str("topic", topic).Msg("Received message from unexpected topic")
	}

	return nil
}

// FlushBuffer는 버퍼에 있는 메시지를 데이터베이스에 저장합니다.
func (c *KafkaConsumer) FlushBuffer() error {
	// 버퍼가 비어있는지 확인
	c.messageBuffer.mu.Lock()
	tracesLen := len(c.messageBuffer.Traces)
	logsLen := len(c.messageBuffer.Logs)
	
	if tracesLen == 0 && logsLen == 0 {
		c.messageBuffer.LastFlushTime = time.Now()
		c.messageBuffer.mu.Unlock()
		return nil
	}

	// 현재 버퍼 내용 복사 후 비우기
	traces := make([]traceDomain.TraceItem, tracesLen)
	logs := make([]logDomain.LogItem, logsLen)
	copy(traces, c.messageBuffer.Traces)
	copy(logs, c.messageBuffer.Logs)
	
	c.messageBuffer.Traces = []traceDomain.TraceItem{}
	c.messageBuffer.Logs = []logDomain.LogItem{}
	c.messageBuffer.LastFlushTime = time.Now()
	c.messageBuffer.mu.Unlock()

	// 트레이스 데이터 저장
	if tracesLen > 0 {
		c.log.Info().Int("count", tracesLen).Msg("Flushing trace data to database")
		err := c.traceService.SaveTraces(traces)
		if err != nil {
			c.log.Error().Err(err).Msg("Error saving traces")
			// 실패 시 다시 버퍼에 추가
			c.messageBuffer.mu.Lock()
			c.messageBuffer.Traces = append(c.messageBuffer.Traces, traces...)
			c.messageBuffer.mu.Unlock()
			return err
		}
	}

	// 로그 데이터 저장
	if logsLen > 0 {
		c.log.Info().Int("count", logsLen).Msg("Flushing log data to database")
		err := c.logService.SaveLogs(logs)
		if err != nil {
			c.log.Error().Err(err).Msg("Error saving logs")
			// 실패 시 다시 버퍼에 추가
			c.messageBuffer.mu.Lock()
			c.messageBuffer.Logs = append(c.messageBuffer.Logs, logs...)
			c.messageBuffer.mu.Unlock()
			return err
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

	// 고루틴 종료 대기
	c.wg.Wait()

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
	// 기존 컨슈머 종료
	err := c.Stop()
	if err != nil {
		c.log.Error().Err(err).Msg("Error stopping Kafka consumer during reconnect")
	}

	// 재연결 전 잠시 대기
	time.Sleep(5 * time.Second)
	
	// 재시작 시도
	c.log.Info().Msg("Attempting to reconnect Kafka consumer")
	retryCount := 0
	for retryCount < 5 {
		retryCount++
		ctx := context.Background()
		if c.ctx != nil {
			ctx = c.ctx
		}
		
		err := c.Start(ctx)
		if err == nil {
			c.log.Info().Msg("Successfully reconnected Kafka consumer")
			return
		}
		
		c.log.Error().Err(err).Int("retry", retryCount).Msg("Failed to reconnect Kafka consumer")
		time.Sleep(time.Duration(retryCount) * 5 * time.Second) // 지수 백오프
	}
	
	c.log.Fatal().Msg("Failed to reconnect Kafka consumer after multiple attempts")
}