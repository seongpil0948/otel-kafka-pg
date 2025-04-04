package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// 필요한 OpenTelemetry 프로토콜 버퍼 정의 URL
var protoURLs = map[string]string{
	"common/v1/common.proto":                         "https://raw.githubusercontent.com/open-telemetry/opentelemetry-proto/main/opentelemetry/proto/common/v1/common.proto",
	"resource/v1/resource.proto":                     "https://raw.githubusercontent.com/open-telemetry/opentelemetry-proto/main/opentelemetry/proto/resource/v1/resource.proto",
	"trace/v1/trace.proto":                           "https://raw.githubusercontent.com/open-telemetry/opentelemetry-proto/main/opentelemetry/proto/trace/v1/trace.proto",
	"logs/v1/logs.proto":                             "https://raw.githubusercontent.com/open-telemetry/opentelemetry-proto/main/opentelemetry/proto/logs/v1/logs.proto",
	"collector/trace/v1/trace_service.proto":         "https://raw.githubusercontent.com/open-telemetry/opentelemetry-proto/main/opentelemetry/proto/collector/trace/v1/trace_service.proto",
	"collector/logs/v1/logs_service.proto":           "https://raw.githubusercontent.com/open-telemetry/opentelemetry-proto/main/opentelemetry/proto/collector/logs/v1/logs_service.proto",
}

func main() {
	// 임시 디렉토리 생성
	tempDir := "temp_proto"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		log.Fatalf("임시 디렉토리 생성 실패: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// proto 출력 디렉토리 생성
	genDir := "proto/gen"
	if err := os.MkdirAll(genDir, 0755); err != nil {
		log.Fatalf("생성 디렉토리 생성 실패: %v", err)
	}

	// 각 proto 파일 다운로드
	for protoPath, url := range protoURLs {
		fullPath := filepath.Join(tempDir, "opentelemetry/proto", protoPath)
		dir := filepath.Dir(fullPath)
		
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("%s 디렉토리 생성 실패: %v", dir, err)
		}

		if err := downloadFile(fullPath, url); err != nil {
			log.Fatalf("%s 다운로드 실패: %v", protoPath, err)
		}
		fmt.Printf("%s 다운로드 완료\n", protoPath)
	}

	// protoc 명령 실행
	protocArgs := []string{
		"--proto_path=" + tempDir,
		"--go_out=" + genDir,
		"--go_opt=paths=source_relative",
		"--go-grpc_out=" + genDir,
		"--go-grpc_opt=paths=source_relative",
	}

	// 모든 proto 파일 추가
	for protoPath := range protoURLs {
		protocArgs = append(protocArgs, filepath.Join("opentelemetry/proto", protoPath))
	}

	cmd := exec.Command("protoc", protocArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("protoc 실행 실패: %v\n%s", err, output)
	}

	fmt.Printf("Protocol Buffer 코드 생성 성공!\n")
	fmt.Printf("생성된 파일들은 %s 디렉토리에 저장됨\n", genDir)

	// import path 수정
	fixImportPaths(genDir)
}

// 파일 다운로드 함수
func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// 생성된 Go 파일의 import 경로 수정
func fixImportPaths(genDir string) {
	err := filepath.Walk(genDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			content, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			// 특정 import 경로 수정
			modified := strings.ReplaceAll(string(content),
				`import "opentelemetry/proto`,
				`import "github.com/seongpil0948/otel-kafka-pg/proto/gen/opentelemetry/proto`)

			// 다른 필요한 경로 수정도 추가 가능

			if modified != string(content) {
				err = ioutil.WriteFile(path, []byte(modified), 0644)
				if err != nil {
					return err
				}
				fmt.Printf("수정됨: %s\n", path)
			}
		}
		return nil
	})

	if err != nil {
		log.Printf("import 경로 수정 중 오류 발생: %v", err)
	} else {
		fmt.Println("import 경로 수정 완료")
	}
}