package config

import (
	"os"
	"strings"
	"sync"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

// Config는 애플리케이션 설정을 관리하는 구조체입니다.
type Config struct {
	Database struct {
		Host     string
		Port     int
		User     string
		Password string
		DBName   string
		MaxConns int
	}

	Kafka struct {
		Brokers       []string
		GroupID       string
		ClientID      string
		TracesTopic   string
		LogsTopic     string
		BatchSize     int
		FlushInterval int
	}

	Logger struct {
		Level string
		IsDev bool
	}
	DataRetention struct {
		Enabled         bool
		CleanupInterval int // 정리 작업 주기(분)
		RetentionPeriod int // 데이터 보존 기간(일)
	}
	API struct {
		Port             int      `json:"port"`
		Host             string   `json:"host"`
		BasePath         string   `json:"basePath"`
		AllowedOrigins   []string `json:"allowedOrigins"`
		AllowCredentials bool     `json:"allowCredentials"`
		ReadTimeout      int      `json:"readTimeout"`
		WriteTimeout     int      `json:"writeTimeout"`
		EnableSwagger    bool     `json:"enableSwagger"`
	}
}

var (
	config *Config
	once   sync.Once
)

// LoadConfig는 환경 변수 또는 파일에서 설정을 로드합니다.
func LoadConfig() *Config {
	once.Do(func() {
		v := viper.New()

		// 환경 변수에서 설정 가져오기
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		v.AutomaticEnv()

		// 기본값 설정
		v.SetDefault("database.host", "localhost")
		v.SetDefault("database.port", 5432)
		v.SetDefault("database.user", "postgres")
		v.SetDefault("database.password", "postgres")
		v.SetDefault("database.dbname", "telemetry")
		v.SetDefault("database.maxconns", 20)

		v.SetDefault("kafka.brokers", []string{"10.101.91.181:9092", "10.101.91.181:9093"})
		v.SetDefault("kafka.groupid", "default-local-group")
		v.SetDefault("kafka.clientid", "default-local-client")
		v.SetDefault("kafka.tracestopic", "onpremise.theshop.oltp.dev.trace")
		v.SetDefault("kafka.logstopic", "onpremise.theshop.oltp.dev.log")
		v.SetDefault("kafka.batchsize", 100)
		v.SetDefault("kafka.flushinterval", 5000)

		v.SetDefault("logger.level", "info")
		v.SetDefault("logger.isdev", false)

		v.SetDefault("dataretention.enabled", true)
		v.SetDefault("dataretention.cleanupinterval", 60) // 1시간 간격 (60분)
		v.SetDefault("dataretention.retentionperiod", 30) // 30일 보존

		v.SetDefault("api.port", 8080)
		v.SetDefault("api.host", "")
		v.SetDefault("api.basePath", "/api")
		v.SetDefault("api.allowedOrigins", []string{"*"})
		v.SetDefault("api.allowCredentials", true)
		v.SetDefault("api.readTimeout", 10)  // 10초
		v.SetDefault("api.writeTimeout", 30) // 30초
		v.SetDefault("api.enableSwagger", true)

		// 환경 변수에서 개별 설정 가져오기
		if host := v.GetString("POSTGRES_HOST"); host != "" {
			v.Set("database.host", host)
		}
		if port := v.GetInt("POSTGRES_PORT"); port != 0 {
			v.Set("database.port", port)
		}
		if user := v.GetString("POSTGRES_USER"); user != "" {
			v.Set("database.user", user)
		}
		if password := v.GetString("POSTGRES_PASSWORD"); password != "" {
			v.Set("database.password", password)
		}
		if dbname := v.GetString("POSTGRES_DB"); dbname != "" {
			v.Set("database.dbname", dbname)
		}
		if maxconns := v.GetInt("POSTGRES_MAX_CONNECTIONS"); maxconns != 0 {
			v.Set("database.maxconns", maxconns)
		}

		// Kafka 설정
		if brokers := v.GetString("KAFKA_BROKERS"); brokers != "" {
			v.Set("kafka.brokers", strings.Split(brokers, ","))
		}
		if groupID := v.GetString("KAFKA_GROUP_ID"); groupID != "" {
			v.Set("kafka.groupid", groupID)
		}
		if clientID := v.GetString("KAFKA_CLIENT_ID"); clientID != "" {
			v.Set("kafka.clientid", clientID)
		}
		if traceTopic := v.GetString("KAFKA_TRACE_TOPIC"); traceTopic != "" {
			v.Set("kafka.tracestopic", traceTopic)
		}
		if logTopic := v.GetString("KAFKA_LOG_TOPIC"); logTopic != "" {
			v.Set("kafka.logstopic", logTopic)
		}
		if batchSize := v.GetInt("BATCH_SIZE"); batchSize != 0 {
			v.Set("kafka.batchsize", batchSize)
		}
		if flushInterval := v.GetInt("FLUSH_INTERVAL"); flushInterval != 0 {
			v.Set("kafka.flushinterval", flushInterval)
		}

		// 로거 설정
		if logLevel := v.GetString("LOG_LEVEL"); logLevel != "" {
			v.Set("logger.level", logLevel)
		}
		if isDev := v.GetString("NODE_ENV"); isDev != "production" {
			v.Set("logger.isdev", true)
		}

		// 데이터 보존 설정
		if enabled := v.GetBool("DATA_RETENTION_ENABLED"); enabled != v.GetBool("dataretention.enabled") {
			v.Set("dataretention.enabled", enabled)
		}
		if interval := v.GetInt("DATA_RETENTION_CLEANUP_INTERVAL"); interval != 0 {
			v.Set("dataretention.cleanupinterval", interval)
		}
		if period := v.GetInt("DATA_RETENTION_PERIOD"); period != 0 {
			v.Set("dataretention.retentionperiod", period)
		}

		if apiPort := v.GetInt("API_PORT"); apiPort != 0 {
			v.Set("api.port", apiPort)
		}

		if apiHost := v.GetString("API_HOST"); apiHost != "" {
			v.Set("api.host", apiHost)
		}

		if apiBasePath := v.GetString("API_BASE_PATH"); apiBasePath != "" {
			v.Set("api.basePath", apiBasePath)
		}

		if origins := v.GetString("API_ALLOWED_ORIGINS"); origins != "" {
			v.Set("api.allowedOrigins", strings.Split(origins, ","))
		}

		if allowCreds := v.GetBool("API_ALLOW_CREDENTIALS"); allowCreds != v.GetBool("api.allowCredentials") {
			v.Set("api.allowCredentials", allowCreds)
		}

		if readTimeout := v.GetInt("API_READ_TIMEOUT"); readTimeout != 0 {
			v.Set("api.readTimeout", readTimeout)
		}

		if writeTimeout := v.GetInt("API_WRITE_TIMEOUT"); writeTimeout != 0 {
			v.Set("api.writeTimeout", writeTimeout)
		}

		if enableSwagger := v.GetBool("API_ENABLE_SWAGGER"); enableSwagger != v.GetBool("api.enableSwagger") {
			v.Set("api.enableSwagger", enableSwagger)
		}

		// 구성 생성
		config = &Config{}

		// 데이터베이스 설정
		config.Database.Host = v.GetString("database.host")
		config.Database.Port = v.GetInt("database.port")
		config.Database.User = v.GetString("database.user")
		config.Database.Password = v.GetString("database.password")
		config.Database.DBName = v.GetString("database.dbname")
		config.Database.MaxConns = v.GetInt("database.maxconns")

		// Kafka 설정
		config.Kafka.Brokers = v.GetStringSlice("kafka.brokers")
		config.Kafka.GroupID = v.GetString("kafka.groupid")
		config.Kafka.ClientID = v.GetString("kafka.clientid")
		config.Kafka.TracesTopic = v.GetString("kafka.tracestopic")
		config.Kafka.LogsTopic = v.GetString("kafka.logstopic")
		config.Kafka.BatchSize = v.GetInt("kafka.batchsize")
		config.Kafka.FlushInterval = v.GetInt("kafka.flushinterval")

		// 로거 설정
		config.Logger.Level = v.GetString("logger.level")
		config.Logger.IsDev = v.GetBool("logger.isdev")

		// 데이터 보존 설정
		config.DataRetention.Enabled = v.GetBool("dataretention.enabled")
		config.DataRetention.CleanupInterval = v.GetInt("dataretention.cleanupinterval")
		config.DataRetention.RetentionPeriod = v.GetInt("dataretention.retentionperiod")

		config.API.Port = v.GetInt("api.port")
		config.API.Host = v.GetString("api.host")
		config.API.BasePath = v.GetString("api.basePath")
		config.API.AllowedOrigins = v.GetStringSlice("api.allowedOrigins")
		config.API.AllowCredentials = v.GetBool("api.allowCredentials")
		config.API.ReadTimeout = v.GetInt("api.readTimeout")
		config.API.WriteTimeout = v.GetInt("api.writeTimeout")
		config.API.EnableSwagger = v.GetBool("api.enableSwagger")

	})

	log := zerolog.New(os.Stdout).With().Timestamp().Logger()
	log.Info().
		Str("database.host", config.Database.Host).
		Int("database.port", config.Database.Port).
		Str("database.user", config.Database.User).
		Str("database.dbname", config.Database.DBName).
		Int("database.maxconns", config.Database.MaxConns).
		Strs("kafka.brokers", config.Kafka.Brokers).
		Str("kafka.groupid", config.Kafka.GroupID).
		Str("kafka.clientid", config.Kafka.ClientID).
		Str("kafka.tracestopic", config.Kafka.TracesTopic).
		Str("kafka.logstopic", config.Kafka.LogsTopic).
		Int("kafka.batchsize", config.Kafka.BatchSize).
		Int("kafka.flushinterval", config.Kafka.FlushInterval).
		Str("logger.level", config.Logger.Level).
		Bool("logger.isdev", config.Logger.IsDev).
		Bool("dataretention.enabled", config.DataRetention.Enabled).
		Int("dataretention.cleanupinterval", config.DataRetention.CleanupInterval).
		Int("dataretention.retentionperiod", config.DataRetention.RetentionPeriod).
		Msg("설정 로드 완료")

	return config
}

// GetConfig는 현재 설정을 반환합니다.
func GetConfig() *Config {
	if config == nil {
		return LoadConfig()
	}
	return config
}
