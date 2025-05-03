package figma

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

const (
	ToolGetFile               = "figma_get_file"
	ToolGetFileNodes          = "figma_get_file_nodes"
	ToolGetFileComponents     = "figma_get_file_components"
	ToolGetFileStyles         = "figma_get_file_styles"
	ToolGetComments           = "figma_get_comments"
	ToolCreateComment         = "figma_create_comment"
	ToolGenerateDesignFromDoc = "figma_generate_design_from_doc"
)

// FigmaClient는 Figma API와 상호작용하는 간단한 클라이언트입니다.
type FigmaClient struct {
	token      string
	httpClient *http.Client
	baseURL    string
}

// NewFigmaClient는 새 Figma 클라이언트를 생성합니다.
func NewFigmaClient(token string) *FigmaClient {
	return &FigmaClient{
		token:      token,
		httpClient: &http.Client{},
		baseURL:    "https://api.figma.com/v1",
	}
}

// isToolEnabledFunc는 도구가 활성화되어 있는지 확인하는 함수 타입입니다.
type isToolEnabledFunc func(string) bool

// RegisterTools는 Figma 관련 도구들을 서버에 등록합니다.
func RegisterTools(
	server *mcpserver.MCPServer,
	token string,
	isToolEnabled isToolEnabledFunc,
) error {
	// 토큰이 없으면 오류
	if token == "" {
		return errors.New("유효한 Figma API 토큰이 필요합니다")
	}

	// Figma 클라이언트 생성
	client := NewFigmaClient(token)

	// 파일 정보 조회 도구 등록
	if isToolEnabled(ToolGetFile) {
		if err := registerGetFileTool(server, client); err != nil {
			return err
		}
	}

	// 파일 노드 조회 도구 등록
	if isToolEnabled(ToolGetFileNodes) {
		if err := registerGetFileNodesTool(server, client); err != nil {
			return err
		}
	}

	// 파일 컴포넌트 조회 도구 등록
	if isToolEnabled(ToolGetFileComponents) {
		if err := registerGetFileComponentsTool(server, client); err != nil {
			return err
		}
	}

	// 파일 스타일 조회 도구 등록
	if isToolEnabled(ToolGetFileStyles) {
		if err := registerGetFileStylesTool(server, client); err != nil {
			return err
		}
	}

	// 코멘트 조회 도구 등록
	if isToolEnabled(ToolGetComments) {
		if err := registerGetCommentsTool(server, client); err != nil {
			return err
		}
	}

	// 코멘트 생성 도구 등록
	if isToolEnabled(ToolCreateComment) {
		if err := registerCreateCommentTool(server, client); err != nil {
			return err
		}
	}

	// 문서 기반 디자인 생성 도구 등록
	if isToolEnabled(ToolGenerateDesignFromDoc) {
		if err := registerGenerateDesignFromDocTool(server, client); err != nil {
			return err
		}
	}

	return nil
}

// registerGetFileTool는 파일 정보 조회 도구를 등록합니다.
func registerGetFileTool(server *mcpserver.MCPServer, client *FigmaClient) error {
	// 파일 정보 조회 도구 정의
	getFileTool := mcp.NewTool(ToolGetFile,
		mcp.WithDescription("Figma 파일 정보 조회"),
		mcp.WithString("file_key",
			mcp.Required(),
			mcp.Description("Figma 파일 키"),
		),
	)

	// 파일 정보 조회 도구 핸들러 등록
	server.AddTool(getFileTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// 이 도구는 실제 구현을 위해 Figma API 호출이 필요합니다.
		// 이 예제에서는 실제 API 호출 없이 샘플 응답만 반환합니다.

		fileKey, ok := request.Params.Arguments["file_key"].(string)
		if !ok || fileKey == "" {
			return mcp.NewToolResultError("유효한 Figma 파일 키가 필요합니다"), nil
		}

		// 샘플 데이터 반환
		result := fmt.Sprintf("Figma 파일 정보 (키: %s):\n\n", fileKey)
		result += "이름: 모바일 앱 디자인 시스템\n"
		result += "마지막 수정: 2024-05-02T14:30:00Z\n"
		result += "버전: 241\n"
		result += "소유자: Jane Doe (jane@example.com)\n\n"
		result += "캔버스:\n"
		result += "- 컴포넌트\n"
		result += "- 스타일 가이드\n"
		result += "- 앱 화면 디자인\n"
		result += "- 로그인 플로우\n"

		return mcp.NewToolResultText(result), nil
	})

	return nil
}

// registerGetFileNodesTool는 파일 노드 조회 도구를 등록합니다.
func registerGetFileNodesTool(server *mcpserver.MCPServer, client *FigmaClient) error {
	// 파일 노드 조회 도구 정의
	getFileNodesTool := mcp.NewTool(ToolGetFileNodes,
		mcp.WithDescription("Figma 파일 노드 조회"),
		mcp.WithString("file_key",
			mcp.Required(),
			mcp.Description("Figma 파일 키"),
		),
		mcp.WithString("node_ids",
			mcp.Required(),
			mcp.Description("조회할 노드 ID (쉼표로 구분)"),
		),
	)

	// 파일 노드 조회 도구 핸들러 등록
	server.AddTool(getFileNodesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// 이 도구는 실제 구현을 위해 Figma API 호출이 필요합니다.
		// 이 예제에서는 실제 API 호출 없이 샘플 응답만 반환합니다.

		fileKey, ok := request.Params.Arguments["file_key"].(string)
		if !ok || fileKey == "" {
			return mcp.NewToolResultError("유효한 Figma 파일 키가 필요합니다"), nil
		}

		nodeIDs, ok := request.Params.Arguments["node_ids"].(string)
		if !ok || nodeIDs == "" {
			return mcp.NewToolResultError("유효한 노드 ID가 필요합니다"), nil
		}

		// 샘플 데이터 반환
		result := fmt.Sprintf("Figma 파일 노드 정보 (파일 키: %s):\n\n", fileKey)

		// 노드 ID 목록을 쉼표로 분리
		nodeIDList := strings.Split(nodeIDs, ",")

		for i, nodeID := range nodeIDList {
			nodeID = strings.TrimSpace(nodeID)

			result += fmt.Sprintf("노드 %d: %s\n", i+1, nodeID)
			result += "  유형: FRAME\n"
			result += "  이름: 로그인 화면\n"
			result += "  크기: 375 x 812 px\n"
			result += "  자식 노드: 12개\n"
			result += "\n"
		}

		return mcp.NewToolResultText(result), nil
	})

	return nil
}

// registerGetFileComponentsTool는 파일 컴포넌트 조회 도구를 등록합니다.
func registerGetFileComponentsTool(server *mcpserver.MCPServer, client *FigmaClient) error {
	// 파일 컴포넌트 조회 도구 정의
	getFileComponentsTool := mcp.NewTool(ToolGetFileComponents,
		mcp.WithDescription("Figma 파일 컴포넌트 조회"),
		mcp.WithString("file_key",
			mcp.Required(),
			mcp.Description("Figma 파일 키"),
		),
	)

	// 파일 컴포넌트 조회 도구 핸들러 등록
	server.AddTool(getFileComponentsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// 이 도구는 실제 구현을 위해 Figma API 호출이 필요합니다.
		// 이 예제에서는 실제 API 호출 없이 샘플 응답만 반환합니다.

		fileKey, ok := request.Params.Arguments["file_key"].(string)
		if !ok || fileKey == "" {
			return mcp.NewToolResultError("유효한 Figma 파일 키가 필요합니다"), nil
		}

		// 샘플 데이터 반환
		result := fmt.Sprintf("Figma 파일 컴포넌트 (파일 키: %s):\n\n", fileKey)

		// 샘플 컴포넌트 목록
		components := []struct {
			Name       string
			Type       string
			Variants   int
			LastUpdate string
		}{
			{"Button", "COMPONENT_SET", 8, "2024-05-01"},
			{"Input Field", "COMPONENT_SET", 6, "2024-04-28"},
			{"Card", "COMPONENT_SET", 4, "2024-04-25"},
			{"Navigation Bar", "COMPONENT", 1, "2024-04-20"},
			{"Tab Bar", "COMPONENT_SET", 5, "2024-04-15"},
		}

		for i, comp := range components {
			result += fmt.Sprintf("%d. %s\n", i+1, comp.Name)
			result += fmt.Sprintf("   유형: %s\n", comp.Type)
			result += fmt.Sprintf("   변형: %d개\n", comp.Variants)
			result += fmt.Sprintf("   마지막 수정: %s\n", comp.LastUpdate)
			result += "\n"
		}

		return mcp.NewToolResultText(result), nil
	})

	return nil
}

// registerGetFileStylesTool는 파일 스타일 조회 도구를 등록합니다.
func registerGetFileStylesTool(server *mcpserver.MCPServer, client *FigmaClient) error {
	// 파일 스타일 조회 도구 정의
	getFileStylesTool := mcp.NewTool(ToolGetFileStyles,
		mcp.WithDescription("Figma 파일 스타일 조회"),
		mcp.WithString("file_key",
			mcp.Required(),
			mcp.Description("Figma 파일 키"),
		),
	)

	// 파일 스타일 조회 도구 핸들러 등록
	server.AddTool(getFileStylesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// 이 도구는 실제 구현을 위해 Figma API 호출이 필요합니다.
		// 이 예제에서는 실제 API 호출 없이 샘플 응답만 반환합니다.

		fileKey, ok := request.Params.Arguments["file_key"].(string)
		if !ok || fileKey == "" {
			return mcp.NewToolResultError("유효한 Figma 파일 키가 필요합니다"), nil
		}

		// 샘플 데이터 반환
		result := fmt.Sprintf("Figma 파일 스타일 (파일 키: %s):\n\n", fileKey)

		// 샘플 스타일 목록
		styles := []struct {
			Name        string
			Type        string
			Description string
		}{
			{"Primary", "색상", "브랜드 주요 색상 (#3366FF)"},
			{"Secondary", "색상", "브랜드 보조 색상 (#FF6633)"},
			{"Heading 1", "텍스트", "32px / Bold / 라인 높이 1.2"},
			{"Heading 2", "텍스트", "24px / SemiBold / 라인 높이 1.3"},
			{"Body", "텍스트", "16px / Regular / 라인 높이 1.5"},
			{"Card Shadow", "이펙트", "8px blur / 30% opacity / y-offset 4px"},
		}

		// 스타일 유형별 분류
		stylesByType := make(map[string][]string)

		for _, style := range styles {
			styleInfo := fmt.Sprintf("%s: %s", style.Name, style.Description)
			stylesByType[style.Type] = append(stylesByType[style.Type], styleInfo)
		}

		// 유형별로 출력
		for styleType, styleInfos := range stylesByType {
			result += fmt.Sprintf("== %s 스타일 ==\n", styleType)
			for i, info := range styleInfos {
				result += fmt.Sprintf("%d. %s\n", i+1, info)
			}
			result += "\n"
		}

		return mcp.NewToolResultText(result), nil
	})

	return nil
}

// registerGetCommentsTool는 코멘트 조회 도구를 등록합니다.
func registerGetCommentsTool(server *mcpserver.MCPServer, client *FigmaClient) error {
	// 코멘트 조회 도구 정의
	getCommentsTool := mcp.NewTool(ToolGetComments,
		mcp.WithDescription("Figma 파일 코멘트 조회"),
		mcp.WithString("file_key",
			mcp.Required(),
			mcp.Description("Figma 파일 키"),
		),
	)

	// 코멘트 조회 도구 핸들러 등록
	server.AddTool(getCommentsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// 이 도구는 실제 구현을 위해 Figma API 호출이 필요합니다.
		// 이 예제에서는 실제 API 호출 없이 샘플 응답만 반환합니다.

		fileKey, ok := request.Params.Arguments["file_key"].(string)
		if !ok || fileKey == "" {
			return mcp.NewToolResultError("유효한 Figma 파일 키가 필요합니다"), nil
		}

		// 샘플 데이터 반환
		result := fmt.Sprintf("Figma 파일 코멘트 (파일 키: %s):\n\n", fileKey)

		// 샘플 코멘트
		comments := []struct {
			ID        string
			Message   string
			Author    string
			CreatedAt string
			Resolved  bool
		}{
			{"123456", "로그인 버튼 색상을 주요 색상으로 변경해주세요.", "Kim Designer", "2024-05-02 15:30", false},
			{"123457", "헤더 텍스트가 모바일에서 너무 큽니다.", "Lee Developer", "2024-05-01 10:15", true},
			{"123458", "이 카드 컴포넌트에 그림자를 추가해주세요.", "Park Manager", "2024-04-30 14:45", false},
		}

		for i, comment := range comments {
			result += fmt.Sprintf("%d. %s\n", i+1, comment.Message)
			result += fmt.Sprintf("   작성자: %s\n", comment.Author)
			result += fmt.Sprintf("   작성일: %s\n", comment.CreatedAt)
			if comment.Resolved {
				result += "   상태: 해결됨\n"
			} else {
				result += "   상태: 미해결\n"
			}
			result += "\n"
		}

		return mcp.NewToolResultText(result), nil
	})

	return nil
}

// registerCreateCommentTool는 코멘트 생성 도구를 등록합니다.
func registerCreateCommentTool(server *mcpserver.MCPServer, client *FigmaClient) error {
	// 코멘트 생성 도구 정의
	createCommentTool := mcp.NewTool(ToolCreateComment,
		mcp.WithDescription("Figma 파일에 코멘트 생성"),
		mcp.WithString("file_key",
			mcp.Required(),
			mcp.Description("Figma 파일 키"),
		),
		mcp.WithString("message",
			mcp.Required(),
			mcp.Description("코멘트 내용"),
		),
		mcp.WithString("node_id",
			mcp.Description("코멘트를 달 노드 ID (선택 사항)"),
		),
	)

	// 코멘트 생성 도구 핸들러 등록
	server.AddTool(createCommentTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// 이 도구는 실제 구현을 위해 Figma API 호출이 필요합니다.
		// 이 예제에서는 실제 API 호출 없이 샘플 응답만 반환합니다.

		fileKey, ok := request.Params.Arguments["file_key"].(string)
		if !ok || fileKey == "" {
			return mcp.NewToolResultError("유효한 Figma 파일 키가 필요합니다"), nil
		}

		message, ok := request.Params.Arguments["message"].(string)
		if !ok || message == "" {
			return mcp.NewToolResultError("유효한 코멘트 내용이 필요합니다"), nil
		}

		nodeID, _ := request.Params.Arguments["node_id"].(string)

		// 샘플 응답 생성
		result := fmt.Sprintf("Figma 코멘트 생성 성공:\n\n")
		result += fmt.Sprintf("파일: %s\n", fileKey)
		result += fmt.Sprintf("코멘트 내용: %s\n", message)

		if nodeID != "" {
			result += fmt.Sprintf("연결된 노드: %s\n", nodeID)
		} else {
			result += "연결된 노드: 없음 (파일 전체 코멘트)\n"
		}

		result += fmt.Sprintf("코멘트 ID: %s\n", generateFakeCommentID())
		result += fmt.Sprintf("생성 시간: %s\n", "2024-05-03 14:30:00")

		return mcp.NewToolResultText(result), nil
	})

	return nil
}

// registerGenerateDesignFromDocTool는 문서 기반 디자인 생성 도구를 등록합니다.
func registerGenerateDesignFromDocTool(server *mcpserver.MCPServer, client *FigmaClient) error {
	// 문서 기반 디자인 생성 도구 정의
	generateDesignTool := mcp.NewTool(ToolGenerateDesignFromDoc,
		mcp.WithDescription("Notion 문서 기반으로 Figma 디자인 생성 (개념적 기능)"),
		mcp.WithString("notion_page_id",
			mcp.Required(),
			mcp.Description("Notion 페이지 ID"),
		),
		mcp.WithString("design_type",
			mcp.Required(),
			mcp.Description("생성할 디자인 유형"),
			mcp.Enum("website", "mobile_app", "dashboard", "presentation"),
		),
		mcp.WithString("style_preferences",
			mcp.Description("스타일 선호도 (쉼표로 구분)"),
		),
	)

	// 문서 기반 디자인 생성 도구 핸들러 등록
	server.AddTool(generateDesignTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// 이 도구는 실제 구현을 위해서는 다양한 API와의 복잡한 통합이 필요합니다.
		// 이 예제에서는 개념적인 응답만 반환합니다.

		notionPageID, ok := request.Params.Arguments["notion_page_id"].(string)
		if !ok || notionPageID == "" {
			return mcp.NewToolResultError("유효한 Notion 페이지 ID가 필요합니다"), nil
		}

		designType, ok := request.Params.Arguments["design_type"].(string)
		if !ok || designType == "" {
			return mcp.NewToolResultError("유효한 디자인 유형이 필요합니다"), nil
		}

		stylePreferences, _ := request.Params.Arguments["style_preferences"].(string)

		// 샘플 응답 생성
		result := fmt.Sprintf("문서 기반 Figma 디자인 생성 요약:\n\n")
		result += fmt.Sprintf("Notion 페이지: %s\n", notionPageID)
		result += fmt.Sprintf("디자인 유형: %s\n", designType)

		if stylePreferences != "" {
			result += fmt.Sprintf("스타일 선호도: %s\n", stylePreferences)
		}

		result += "\n이 도구는 개념적인 기능으로, 실제 구현에는 다음 단계가 필요합니다:\n\n"
		result += "1. Notion API를 사용하여 문서 내용 가져오기\n"
		result += "2. 문서 구조 및 내용 분석\n"
		result += "3. 디자인 패턴 및 구조 결정\n"
		result += "4. Figma API를 사용하여 디자인 요소 생성\n"
		result += "5. 텍스트 및 이미지 배치\n"
		result += "6. 스타일 적용\n\n"

		result += "이러한 과정은 LLM과 같은 추가 기술을 활용하여 더 지능적으로 수행할 수 있습니다.\n"
		result += "현재는 실제 파일 생성 없이 개념적인 응답만 제공합니다.\n\n"

		result += "가상의 생성된 Figma 파일:\n"
		result += "파일 키: XYZ123456789\n"
		result += "파일 링크: https://www.figma.com/file/XYZ123456789\n"

		return mcp.NewToolResultText(result), nil
	})

	return nil
}

// generateFakeCommentID는 테스트용 가짜 코멘트 ID를 생성합니다.
func generateFakeCommentID() string {
	return "comment123456789"
}
