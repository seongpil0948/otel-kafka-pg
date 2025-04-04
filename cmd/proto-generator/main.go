package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

// 필요한 proto 파일 목록
var protoFiles = []string{
	"opentelemetry/proto/common/v1/common.proto",
	"opentelemetry/proto/resource/v1/resource.proto",
	"opentelemetry/proto/trace/v1/trace.proto",
	"opentelemetry/proto/logs/v1/logs.proto",
	"opentelemetry/proto/metrics/v1/metrics.proto",
}

// GitHub raw 콘텐츠 URL
const githubRawBase = "https://raw.githubusercontent.com/open-telemetry/opentelemetry-proto/main/"

// 출력 디렉토리
const (
	outputDir      = "./proto"
	tempProtoDir   = "./temp_proto"
	generatedGoDir = "./proto/gen"
)

func main() {
	// 디렉토리 생성
	createDirectories()

	// proto 파일 다운로드
	downloadProtoFiles()

	// protoc 명령어 실행
	generateGoCode()

	// 임시 디렉토리 정리
	cleanupTempDir()

	fmt.Println("Proto 파일 생성 완료!")
}

// 필요한 디렉토리 생성
func createDirectories() {
	dirs := []string{outputDir, tempProtoDir, generatedGoDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("디렉토리 생성 실패 %s: %v", dir, err)
		}
	}
}

// proto 파일 다운로드
func downloadProtoFiles() {
	for _, file := range protoFiles {
		// 파일 경로 생성
		fullPath := filepath.Join(tempProtoDir, file)
		dirPath := filepath.Dir(fullPath)

		// 디렉토리가 없으면 생성
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			log.Fatalf("디렉토리 생성 실패 %s: %v", dirPath, err)
		}

		// 파일 URL 생성
		fileURL := githubRawBase + file
		fmt.Printf("다운로드 중: %s\n", fileURL)

		// HTTP 요청
		resp, err := http.Get(fileURL)
		if err != nil {
			log.Fatalf("파일 다운로드 실패 %s: %v", file, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Fatalf("파일 다운로드 실패 %s: 상태 코드 %d", file, resp.StatusCode)
		}

		// 파일 생성
		out, err := os.Create(fullPath)
		if err != nil {
			log.Fatalf("파일 생성 실패 %s: %v", fullPath, err)
		}
		defer out.Close()

		// 파일 복사
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			log.Fatalf("파일 저장 실패 %s: %v", fullPath, err)
		}

		fmt.Printf("다운로드 완료: %s\n", file)
	}
}

// Go 코드 생성
func generateGoCode() {
	// protoc 명령어가 설치되어 있는지 확인
	if _, err := exec.LookPath("protoc"); err != nil {
		log.Fatalf("protoc가 설치되어 있지 않습니다. 설치 후 다시 시도해주세요.")
	}

	// protoc-gen-go가 설치되어 있는지 확인
	if _, err := exec.LookPath("protoc-gen-go"); err != nil {
		log.Fatalf("protoc-gen-go가 설치되어 있지 않습니다. 'go install google.golang.org/protobuf/cmd/protoc-gen-go@latest'로 설치해주세요.")
	}

	// protoc-gen-go-grpc가 설치되어 있는지 확인
	if _, err := exec.LookPath("protoc-gen-go-grpc"); err != nil {
		log.Fatalf("protoc-gen-go-grpc가 설치되어 있지 않습니다. 'go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest'로 설치해주세요.")
	}

	fmt.Println("Go 코드 생성 중...")

	// protoc 명령어 실행
	args := []string{
		"--proto_path=" + tempProtoDir,
		"--go_out=" + generatedGoDir,
		"--go_opt=paths=source_relative",
		"--go-grpc_out=" + generatedGoDir,
		"--go-grpc_opt=paths=source_relative",
	}

	// 모든 proto 파일 추가
	for _, file := range protoFiles {
		args = append(args, filepath.Join(tempProtoDir, file))
	}

	cmd := exec.Command("protoc", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("protoc 실행 실패: %v\n%s", err, output)
	}

	fmt.Println("Go 코드 생성 완료!")
}

// 임시 디렉토리 정리
func cleanupTempDir() {
	fmt.Println("임시 파일 정리 중...")
	if err := os.RemoveAll(tempProtoDir); err != nil {
		log.Printf("임시 디렉토리 삭제 실패: %v", err)
	}
}