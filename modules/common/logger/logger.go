package logger

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/seongpil0948/otel-kafka-pg/modules/common/config"
)

// Logger는 애플리케이션에서 사용할 로거 인터페이스입니다.
type Logger interface {
	Debug() *zerolog.Event
	Info() *zerolog.Event
	Warn() *zerolog.Event
	Error() *zerolog.Event
	Fatal() *zerolog.Event
	Panic() *zerolog.Event
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
}

// ZeroLogger는 zerolog 구현체입니다.
type ZeroLogger struct {
	logger zerolog.Logger
}

var defaultLogger *ZeroLogger

// 기본 로그 레벨 설정
var logLevel = zerolog.InfoLevel

// Init은 로거를 초기화합니다.
func Init() Logger {
	cfg := config.GetConfig()
	
	// 환경 변수에서 로그 레벨 가져오기
	configLogLevel := strings.ToLower(cfg.Logger.Level)
	switch configLogLevel {
	case "trace":
		logLevel = zerolog.TraceLevel
	case "debug":
		logLevel = zerolog.DebugLevel
	case "info":
		logLevel = zerolog.InfoLevel
	case "warn":
		logLevel = zerolog.WarnLevel
	case "error":
		logLevel = zerolog.ErrorLevel
	case "fatal":
		logLevel = zerolog.FatalLevel
	case "panic":
		logLevel = zerolog.PanicLevel
	}

	// 개발 환경에서는 콘솔 형식으로 출력, 프로덕션에서는 JSON 형식으로 출력
	var zl zerolog.Logger
	
	if cfg.Logger.IsDev {
		output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05"}
		zl = zerolog.New(output).With().Timestamp().Caller().Logger().Level(logLevel)
	} else {
		zl = zerolog.New(os.Stdout).With().Timestamp().Logger().Level(logLevel)
	}

	// 기본 로거 설정
	log.Logger = zl
	defaultLogger = &ZeroLogger{logger: zl}
	
	return defaultLogger
}

// GetLogger는 현재 로거 인스턴스를 반환합니다.
func GetLogger() Logger {
	if defaultLogger == nil {
		return Init()
	}
	return defaultLogger
}

// Debug는 디버그 레벨 로깅을 위한 이벤트를 반환합니다.
func (l *ZeroLogger) Debug() *zerolog.Event {
	return l.logger.Debug()
}

// Info는 정보 레벨 로깅을 위한 이벤트를 반환합니다.
func (l *ZeroLogger) Info() *zerolog.Event {
	return l.logger.Info()
}

// Warn은 경고 레벨 로깅을 위한 이벤트를 반환합니다.
func (l *ZeroLogger) Warn() *zerolog.Event {
	return l.logger.Warn()
}

// Error는 오류 레벨 로깅을 위한 이벤트를 반환합니다.
func (l *ZeroLogger) Error() *zerolog.Event {
	return l.logger.Error()
}

// Fatal은 치명적 레벨 로깅을 위한 이벤트를 반환합니다.
func (l *ZeroLogger) Fatal() *zerolog.Event {
	return l.logger.Fatal()
}

// Panic은 패닉 레벨 로깅을 위한 이벤트를 반환합니다.
func (l *ZeroLogger) Panic() *zerolog.Event {
	return l.logger.Panic()
}

// WithField는 필드가 추가된 새 로거를 반환합니다.
func (l *ZeroLogger) WithField(key string, value interface{}) Logger {
	return &ZeroLogger{
		logger: l.logger.With().Interface(key, value).Logger(),
	}
}

// WithFields는 여러 필드가 추가된 새 로거를 반환합니다.
func (l *ZeroLogger) WithFields(fields map[string]interface{}) Logger {
	ctx := l.logger.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	return &ZeroLogger{
		logger: ctx.Logger(),
	}
}
