package server

import (
	"context"
	"errors"
	"fmt"

	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/seongpil0948/otel-kafka-pg/modules/common/logger"

	"github.com/seongpil0948/otel-kafka-pg/modules/mcp/tools/figma"
	"github.com/seongpil0948/otel-kafka-pg/modules/mcp/tools/gitlab"
	"github.com/seongpil0948/otel-kafka-pg/modules/mcp/tools/notion"
	"github.com/seongpil0948/otel-kafka-pg/modules/mcp/tools/telemetry"
)

// MCPServer는 MCP 서버를 관리하는 구조체입니다.
type MCPServer struct {
	server       *mcpserver.MCPServer
	config       *Config
	log          logger.Logger
	toolRegistry map[string]bool // 등록된 도구 추적
}

// NewMCPServer는 새 MCP 서버 인스턴스를 생성합니다.
func NewMCPServer(config *Config, log logger.Logger) (*MCPServer, error) {
	if config == nil {
		config = DefaultConfig()
	}

	srv := mcpserver.NewMCPServer(
		"OpenTelemetry MCP Server",
		"1.0.0",
	)

	// 로그 설정이 활성화된 경우 로깅 옵션 적용
	if config.LogEnabled {
		// 커스텀 로거 적용할 수 있으면 추가
	}

	return &MCPServer{
		server:       srv,
		config:       config,
		log:          log,
		toolRegistry: make(map[string]bool),
	}, nil
}

// RegisterTools는 서버에 도구들을 등록합니다.
func (s *MCPServer) RegisterTools(ctx context.Context) error {
	s.log.Info().Msg("MCP 도구 등록 시작")

	// 등록 오류 수집
	var errs []error

	// 도구 등록 함수 실행
	if err := s.registerGitLabTools(); err != nil {
		errs = append(errs, fmt.Errorf("GitLab 도구 등록 실패: %w", err))
	}

	if err := s.registerNotionTools(); err != nil {
		errs = append(errs, fmt.Errorf("Notion 도구 등록 실패: %w", err))
	}

	if err := s.registerFigmaTools(); err != nil {
		errs = append(errs, fmt.Errorf("Figma 도구 등록 실패: %w", err))
	}

	if err := s.registerTelemetryTools(); err != nil {
		errs = append(errs, fmt.Errorf("텔레메트리 도구 등록 실패: %w", err))
	}

	// 등록된 도구 수 로깅
	s.log.Info().Int("registeredToolCount", len(s.toolRegistry)).Msg("MCP 도구 등록 완료")

	// 오류가 있으면 병합하여 반환
	if len(errs) > 0 {
		errMsg := "도구 등록 중 오류 발생: "
		for i, err := range errs {
			if i > 0 {
				errMsg += "; "
			}
			errMsg += err.Error()
		}
		return errors.New(errMsg)
	}

	return nil
}

// Start는 MCP 서버를 시작합니다.
func (s *MCPServer) Start(ctx context.Context) error {
	s.log.Info().Msg("MCP 서버 시작")
	return mcpserver.ServeStdio(s.server)
}

// isToolEnabled는 도구가 사용 가능한지 확인합니다.
func (s *MCPServer) isToolEnabled(toolName string) bool {
	// 설정된 사용 가능 도구가 없으면 모든 도구 활성화
	if len(s.config.EnabledTools) == 0 {
		return true
	}

	// 도구 이름으로 활성화 여부 확인
	for _, enabled := range s.config.EnabledTools {
		if enabled == toolName {
			return true
		}
	}

	return false
}

// registerGitLabTools는 GitLab 관련 도구를 등록합니다.
func (s *MCPServer) registerGitLabTools() error {
	// GitLab 토큰이 없으면 건너뜀
	if s.config.GitLabToken == "" {
		s.log.Warn().Msg("GitLab 토큰이 설정되지 않아 GitLab 도구를 등록하지 않습니다")
		return nil
	}

	// GitLab 도구 등록
	return gitlab.RegisterTools(s.server, s.config.GitLabToken, s.config.GitLabURL, s.isToolEnabled)
}

// registerNotionTools는 Notion 관련 도구를 등록합니다.
func (s *MCPServer) registerNotionTools() error {
	// Notion 토큰이 없으면 건너뜀
	if s.config.NotionToken == "" {
		s.log.Warn().Msg("Notion 토큰이 설정되지 않아 Notion 도구를 등록하지 않습니다")
		return nil
	}

	// Notion 도구 등록
	return notion.RegisterTools(s.server, s.config.NotionToken, s.isToolEnabled)
}

// registerFigmaTools는 Figma 관련 도구를 등록합니다.
func (s *MCPServer) registerFigmaTools() error {
	// Figma 토큰이 없으면 건너뜀
	if s.config.FigmaToken == "" {
		s.log.Warn().Msg("Figma 토큰이 설정되지 않아 Figma 도구를 등록하지 않습니다")
		return nil
	}

	// Figma 도구 등록
	return figma.RegisterTools(s.server, s.config.FigmaToken, s.isToolEnabled)
}

// registerTelemetryTools는 텔레메트리 관련 도구를 등록합니다.
func (s *MCPServer) registerTelemetryTools() error {
	// 텔레메트리 도구 항상 등록 시도
	return telemetry.RegisterTools(s.server, s.isToolEnabled)
}
