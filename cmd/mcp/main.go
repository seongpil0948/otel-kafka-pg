package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/seongpil0948/otel-kafka-pg/modules/common/logger"
	"github.com/seongpil0948/otel-kafka-pg/modules/mcp/server"
)

// 명령행 인자
var (
	configPath   = flag.String("config", "", "설정 파일 경로 (기본값: 환경 변수 사용)")
	gitlabToken  = flag.String("gitlab-token", "", "GitLab API 토큰 (기본값: 환경 변수 GITLAB_TOKEN)")
	gitlabURL    = flag.String("gitlab-url", "", "GitLab 서버 URL (기본값: 환경 변수 GITLAB_URL)")
	notionToken  = flag.String("notion-token", "", "Notion API 토큰 (기본값: 환경 변수 NOTION_TOKEN)")
	figmaToken   = flag.String("figma-token", "", "Figma API 토큰 (기본값: 환경 변수 FIGMA_TOKEN)")
	enabledTools = flag.String("enabled-tools", "", "사용할 도구 목록 (쉼표로 구분, 기본값: 모두 사용)")
	debug        = flag.Bool("debug", false, "디버그 모드 활성화")
)

func main() {
	// 명령행 인자 파싱
	flag.Parse()

	// 컨텍스트 생성
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// cfg := config.LoadConfig()

	// 로거 초기화
	log := logger.Init()
	log.Info().Msg("OpenTelemetry MCP 서버 시작 중...")

	// 환경 변수에서 토큰 가져오기 (명령행 인자로 지정되지 않은 경우)
	if *gitlabToken == "" {
		*gitlabToken = os.Getenv("GITLAB_TOKEN")
	}

	if *gitlabURL == "" {
		*gitlabURL = os.Getenv("GITLAB_URL")
	}

	if *notionToken == "" {
		*notionToken = os.Getenv("NOTION_TOKEN")
	}

	if *figmaToken == "" {
		*figmaToken = os.Getenv("FIGMA_TOKEN")
	}

	// 로그 레벨 설정
	if *debug {
		log.Info().Msg("디버그 모드가 활성화되었습니다.")
	}

	// 설정된 도구 목록 파싱
	var enabledToolsList []string
	if *enabledTools != "" {
		enabledToolsList = parseCommaSeparatedList(*enabledTools)
	}

	// MCP 서버 설정
	mcpConfig := &server.Config{
		GitLabToken:  *gitlabToken,
		GitLabURL:    *gitlabURL,
		NotionToken:  *notionToken,
		FigmaToken:   *figmaToken,
		LogEnabled:   *debug,
		EnabledTools: enabledToolsList,
	}

	// MCP 서버 생성
	mcpServer, err := server.NewMCPServer(mcpConfig, log)
	if err != nil {
		log.Fatal().Err(err).Msg("MCP 서버 생성 실패")
	}

	// 도구 등록
	if err := mcpServer.RegisterTools(ctx); err != nil {
		log.Fatal().Err(err).Msg("MCP 도구 등록 실패")
	}

	// 종료 시그널 처리 설정
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	// 종료 시그널 처리 고루틴
	go func() {
		sig := <-sigCh
		log.Info().Str("signal", sig.String()).Msg("종료 신호 수신, 정상 종료를 시작합니다")
		cancel() // 컨텍스트 취소하여 모든 작업 종료
	}()

	// 서버 시작
	log.Info().Msg("MCP 서버 시작 중...")
	if err := mcpServer.Start(ctx); err != nil {
		log.Fatal().Err(err).Msg("MCP 서버 시작 실패")
	}
}

// parseCommaSeparatedList는 쉼표로 구분된 문자열을 문자열 슬라이스로 변환합니다.
func parseCommaSeparatedList(input string) []string {
	if input == "" {
		return nil
	}

	return splitAndTrim(input, ",")
}

// splitAndTrim은 지정된 구분자로 문자열을 분할하고 공백을 제거합니다.
func splitAndTrim(input, separator string) []string {
	if input == "" {
		return nil
	}

	parts := []string{}
	for _, part := range strings.Split(input, separator) {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}

	return parts
}
