package db

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"github.com/seongpil0948/otel-kafka-pg/modules/common/config"
	"github.com/seongpil0948/otel-kafka-pg/modules/common/logger"
)

// Database는 데이터베이스 인터페이스입니다.
type Database interface {
	GetDB() *sql.DB
	Close() error
	Begin() (*sql.Tx, error)
	Execute(query string, args ...interface{}) (sql.Result, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

// PostgresDB는 PostgreSQL 구현체입니다.
type PostgresDB struct {
	db *sql.DB
}

var (
	instance *PostgresDB
	once     sync.Once
	log      = logger.GetLogger()
)

// NewDatabase는 새 데이터베이스 인스턴스를 생성합니다.
func NewDatabase() (Database, error) {
	var err error

	once.Do(func() {
		cfg := config.GetConfig()

		// 연결 문자열 생성
		connStr := fmt.Sprintf(
			"user=%s host=%s dbname=%s password=%s port=%d sslmode=disable",
			cfg.Database.User,
			cfg.Database.Host,
			cfg.Database.DBName,
			cfg.Database.Password,
			cfg.Database.Port,
		)

		db, dbErr := sql.Open("postgres", connStr)
		if dbErr != nil {
			err = fmt.Errorf("데이터베이스 연결 실패: %w", dbErr)
			return
		}

		// 설정
		db.SetMaxOpenConns(cfg.Database.MaxConns)
		db.SetMaxIdleConns(10)
		db.SetConnMaxLifetime(time.Minute * 5)
		db.SetConnMaxIdleTime(time.Second * 30)

		// 연결 테스트
		if pingErr := db.Ping(); pingErr != nil {
			err = fmt.Errorf("데이터베이스 ping 실패: %w", pingErr)
			return
		}

		instance = &PostgresDB{db: db}
		log.Info().
			Str("host", cfg.Database.Host).
			Int("port", cfg.Database.Port).
			Str("dbname", cfg.Database.DBName).
			Msg("데이터베이스 연결 성공")
	})

	if err != nil {
		return nil, err
	}

	return instance, nil
}

// GetInstance는 싱글톤 인스턴스를 반환합니다.
func GetInstance() (Database, error) {
	if instance == nil {
		return NewDatabase()
	}
	return instance, nil
}

// GetDB는 내부 데이터베이스 객체를 반환합니다.
func (p *PostgresDB) GetDB() *sql.DB {
	return p.db
}

// Close는 데이터베이스 연결을 종료합니다.
func (p *PostgresDB) Close() error {
	log.Info().Msg("데이터베이스 연결 종료 중...")
	err := p.db.Close()
	if err != nil {
		return fmt.Errorf("데이터베이스 연결 종료 실패: %w", err)
	}
	log.Info().Msg("데이터베이스 연결 종료됨")
	return nil
}

// Begin은 새 트랜잭션을 시작합니다.
func (p *PostgresDB) Begin() (*sql.Tx, error) {
	return p.db.Begin()
}

// Execute는 SQL 쿼리를 실행합니다.
func (p *PostgresDB) Execute(query string, args ...interface{}) (sql.Result, error) {
	return p.db.Exec(query, args...)
}

// QueryRow는 단일 행을 쿼리합니다.
func (p *PostgresDB) QueryRow(query string, args ...interface{}) *sql.Row {
	return p.db.QueryRow(query, args...)
}

// Query는 여러 행을 쿼리합니다.
func (p *PostgresDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return p.db.Query(query, args...)
}
