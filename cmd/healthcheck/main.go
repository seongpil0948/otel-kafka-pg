package main

import (
	"fmt"
	"os"

	"github.com/seongpil0948/otel-kafka-pg/modules/common/config"
	"github.com/seongpil0948/otel-kafka-pg/modules/common/db"
)

func main() {
	// 설정 로드 (사용하지 않으므로 변수 할당 제거)
	_ = config.LoadConfig()
	
	fmt.Println("Health check 실행 중...")
	
	// 데이터베이스 연결 확인
	database, err := db.NewDatabase()
	if err != nil {
		fmt.Printf("Health check 실패: 데이터베이스 연결 오류: %v\n", err)
		os.Exit(1)
	}
	
	// 간단한 쿼리 실행
	_, err = database.Execute("SELECT 1")
	if err != nil {
		fmt.Printf("Health check 실패: 데이터베이스 쿼리 오류: %v\n", err)
		os.Exit(1)
	}
	
	// 데이터베이스 연결 종료
	if err := database.Close(); err != nil {
		fmt.Printf("Health check 실패: 데이터베이스 연결 종료 오류: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("Health check 성공")
	os.Exit(0)
}