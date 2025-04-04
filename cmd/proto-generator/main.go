package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

// 생성할 OpenTelemetry Proto 파일 목록
var protoFiles = []string{
	"opentelemetry/proto/common/v1/common.proto",
	"opentelemetry/proto/resource/v1/resource.proto",
	"opentelemetry/proto/trace/v1/trace.proto",
	"opentelemetry/proto/logs/v1/logs.proto",
	"opentelemetry/proto/collector/trace/v1/trace_service.proto",
	"opentelemetry/proto/collector/logs/v1/logs_service.proto",
}

// 베이스 URL
const baseURL = "https://raw.githubusercontent.com/open-telemetry/opentelemetry-proto/main"

func main() {
	fmt.Println("OpenTelemetry proto 파일 생성기 시작...")

	// 작업 디렉토리 설정
	workDir, err := filepath.Abs(".")
	if err != nil {
		fmt.Printf("작업 디렉토리 확인 실패: %v\n", err)
		os.Exit(1)
	}

	// 임시 디렉토리 생성
	tempDir := filepath.Join(workDir, "temp_proto")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		fmt.Printf("임시 디렉토리 생성 실패: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tempDir)

	// 출력 디렉토리 생성
	outDir := filepath.Join(workDir, "..", "..", "proto", "gen")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Printf("출력 디렉토리 생성 실패: %v\n", err)
		os.Exit(1)
	}

	// Proto 파일 다운로드
	for _, protoFile := range protoFiles {
		downloadPath := filepath.Join(tempDir, protoFile)
		dir := filepath.Dir(downloadPath)
		
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("디렉토리 생성 실패 %s: %v\n", dir, err)
			continue
		}

		url := fmt.Sprintf("%s/%s", baseURL, protoFile)
		fmt.Printf("다운로드 중: %s\n", url)
		
		if err := downloadFile(url, downloadPath); err != nil {
			fmt.Printf("다운로드 실패 %s: %v\n", protoFile, err)
			continue
		}
	}

	// protoc 명령 실행
	fmt.Println("protoc 실행 중...")
	
	protocArgs := []string{
		fmt.Sprintf("--proto_path=%s", tempDir),
		fmt.Sprintf("--go_out=%s", outDir),
		fmt.Sprintf("--go-grpc_out=%s", outDir),
	}
	
	// 모든 proto 파일 추가
	for _, protoFile := range protoFiles {
		protocArgs = append(protocArgs, filepath.Join(tempDir, protoFile))
	}

	cmd := exec.Command("protoc", protocArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("protoc 실행 실패: %v\n", err)
		fmt.Println(string(output))
		os.Exit(1)
	}

	fmt.Println("성공적으로 프로토콜 버퍼 파일 생성 완료!")
	fmt.Printf("출력 디렉토리: %s\n", outDir)

	// Go 모듈 생성
	createGoModFile(filepath.Join(workDir, "..", "..", "proto"))
}

// 파일 다운로드 함수
func downloadFile(url, filePath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP 오류: %s", resp.Status)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

// go.mod 파일 생성 함수
func createGoModFile(protoDir string) error {
	goModPath := filepath.Join(protoDir, "go.mod")
	
	// 이미 존재하는지 확인
	if _, err := os.Stat(goModPath); err == nil {
		fmt.Println("go.mod 파일이 이미 존재합니다. 건너뜁니다.")
		return nil
	}

	content := `module github.com/seongpil0948/otel-kafka-pg/proto

go 1.24

require (
	google.golang.org/grpc v1.62.1
	google.golang.org/protobuf v1.33.0
)
`

	return os.WriteFile(goModPath, []byte(content), 0644)
}