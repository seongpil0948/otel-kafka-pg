package notion

import (
	"context"
	"errors"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

const (
	ToolSearchPages   = "notion_search_pages"
	ToolGetPage       = "notion_get_page"
	ToolCreatePage    = "notion_create_page"
	ToolGetDatabases  = "notion_get_databases"
	ToolQueryDatabase = "notion_query_database"
)

// NotionClient는 Notion API와 상호작용하는 간단한 클라이언트입니다.
type NotionClient struct {
	token string
}

// NewNotionClient는 새 Notion 클라이언트를 생성합니다.
func NewNotionClient(token string) *NotionClient {
	return &NotionClient{
		token: token,
	}
}

// isToolEnabledFunc는 도구가 활성화되어 있는지 확인하는 함수 타입입니다.
type isToolEnabledFunc func(string) bool

// RegisterTools는 Notion 관련 도구들을 서버에 등록합니다.
func RegisterTools(
	server *mcpserver.MCPServer,
	token string,
	isToolEnabled isToolEnabledFunc,
) error {
	// 토큰이 없으면 오류
	if token == "" {
		return errors.New("유효한 Notion API 토큰이 필요합니다")
	}

	// Notion 클라이언트 생성
	client := NewNotionClient(token)

	// 페이지 검색 도구 등록
	if isToolEnabled(ToolSearchPages) {
		if err := registerSearchPagesTool(server, client); err != nil {
			return err
		}
	}

	// 페이지 조회 도구 등록
	if isToolEnabled(ToolGetPage) {
		if err := registerGetPageTool(server, client); err != nil {
			return err
		}
	}

	// 페이지 생성 도구 등록
	if isToolEnabled(ToolCreatePage) {
		if err := registerCreatePageTool(server, client); err != nil {
			return err
		}
	}

	// 데이터베이스 목록 조회 도구 등록
	if isToolEnabled(ToolGetDatabases) {
		if err := registerGetDatabasesTool(server, client); err != nil {
			return err
		}
	}

	// 데이터베이스 쿼리 도구 등록
	if isToolEnabled(ToolQueryDatabase) {
		if err := registerQueryDatabaseTool(server, client); err != nil {
			return err
		}
	}

	return nil
}

// registerSearchPagesTool는 페이지 검색 도구를 등록합니다.
func registerSearchPagesTool(server *mcpserver.MCPServer, client *NotionClient) error {
	// 페이지 검색 도구 정의
	searchPagesTool := mcp.NewTool(ToolSearchPages,
		mcp.WithDescription("Notion 페이지 검색"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("검색어"),
		),
		mcp.WithNumber("limit",
			mcp.Description("결과 제한 수"),
		),
	)

	// 페이지 검색 도구 핸들러 등록
	server.AddTool(searchPagesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// 이 도구는 실제 구현을 위해 Notion API 호출이 필요합니다.
		// 이 예제에서는 실제 API 호출 없이 샘플 응답만 반환합니다.

		query, ok := request.Params.Arguments["query"].(string)
		if !ok || query == "" {
			return mcp.NewToolResultError("유효한 검색어가 필요합니다"), nil
		}

		limit := 5 // 기본값
		if limitVal, ok := request.Params.Arguments["limit"].(float64); ok {
			limit = int(limitVal)
		}

		// 샘플 데이터 반환
		result := fmt.Sprintf("Notion 검색 결과 - 쿼리: '%s', 제한: %d개\n\n", query, limit)
		result += "1. 개발 가이드라인 (페이지)\n"
		result += "   URL: https://www.notion.so/company/123456789abcd\n"
		result += "   마지막 편집: 2024-05-01\n\n"
		result += "2. 기술 스택 정리 (데이터베이스)\n"
		result += "   URL: https://www.notion.so/company/987654321dcba\n"
		result += "   마지막 편집: 2024-04-15\n\n"

		return mcp.NewToolResultText(result), nil
	})

	return nil
}

// registerGetPageTool는 페이지 조회 도구를 등록합니다.
func registerGetPageTool(server *mcpserver.MCPServer, client *NotionClient) error {
	// 페이지 조회 도구 정의
	getPageTool := mcp.NewTool(ToolGetPage,
		mcp.WithDescription("Notion 페이지 상세 정보 조회"),
		mcp.WithString("page_id",
			mcp.Required(),
			mcp.Description("Notion 페이지 ID"),
		),
	)

	// 페이지 조회 도구 핸들러 등록
	server.AddTool(getPageTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// 이 도구는 실제 구현을 위해 Notion API 호출이 필요합니다.
		// 이 예제에서는 실제 API 호출 없이 샘플 응답만 반환합니다.

		pageID, ok := request.Params.Arguments["page_id"].(string)
		if !ok || pageID == "" {
			return mcp.NewToolResultError("유효한 페이지 ID가 필요합니다"), nil
		}

		// 샘플 데이터 반환
		result := fmt.Sprintf("Notion 페이지 정보 (ID: %s)\n\n", pageID)
		result += "제목: API 개발 가이드라인\n\n"
		result += "내용:\n"
		result += "# API 개발 가이드라인\n\n"
		result += "이 문서는 회사의 API 개발 표준 가이드라인을 제공합니다.\n\n"
		result += "## 기본 원칙\n\n"
		result += "1. RESTful 디자인 원칙 준수\n"
		result += "2. 명확한 에러 처리\n"
		result += "3. 적절한 문서화\n\n"
		result += "## 인증 방식\n\n"
		result += "모든 API는 JWT를 사용한 인증을 지원해야 합니다.\n"

		return mcp.NewToolResultText(result), nil
	})

	return nil
}

// registerCreatePageTool는 페이지 생성 도구를 등록합니다.
func registerCreatePageTool(server *mcpserver.MCPServer, client *NotionClient) error {
	// 페이지 생성 도구 정의
	createPageTool := mcp.NewTool(ToolCreatePage,
		mcp.WithDescription("Notion 페이지 생성"),
		mcp.WithString("parent_id",
			mcp.Required(),
			mcp.Description("상위 페이지 또는 데이터베이스 ID"),
		),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("페이지 제목"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("페이지 내용 (마크다운 형식)"),
		),
	)

	// 페이지 생성 도구 핸들러 등록
	server.AddTool(createPageTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// 이 도구는 실제 구현을 위해 Notion API 호출이 필요합니다.
		// 이 예제에서는 실제 API 호출 없이 샘플 응답만 반환합니다.

		parentID, ok := request.Params.Arguments["parent_id"].(string)
		if !ok || parentID == "" {
			return mcp.NewToolResultError("유효한 상위 페이지 또는 데이터베이스 ID가 필요합니다"), nil
		}

		title, ok := request.Params.Arguments["title"].(string)
		if !ok || title == "" {
			return mcp.NewToolResultError("유효한 페이지 제목이 필요합니다"), nil
		}

		content, ok := request.Params.Arguments["content"].(string)
		if !ok || content == "" {
			return mcp.NewToolResultError("유효한 페이지 내용이 필요합니다"), nil
		}

		// 샘플 응답 반환
		result := fmt.Sprintf("Notion 페이지가 성공적으로 생성되었습니다.\n\n")
		result += fmt.Sprintf("제목: %s\n", title)
		result += fmt.Sprintf("상위 페이지/데이터베이스: %s\n", parentID)
		result += fmt.Sprintf("내용 미리보기: %s...\n", content[:min(100, len(content))])
		result += fmt.Sprintf("URL: https://www.notion.so/workspace/%s\n", generateFakeID())

		return mcp.NewToolResultText(result), nil
	})

	return nil
}

// registerGetDatabasesTool는 데이터베이스 목록 조회 도구를 등록합니다.
func registerGetDatabasesTool(server *mcpserver.MCPServer, client *NotionClient) error {
	// 데이터베이스 목록 조회 도구 정의
	getDatabasesTool := mcp.NewTool(ToolGetDatabases,
		mcp.WithDescription("Notion 데이터베이스 목록 조회"),
		mcp.WithNumber("limit",
			mcp.Description("결과 제한 수"),
		),
	)

	// 데이터베이스 목록 조회 도구 핸들러 등록
	server.AddTool(getDatabasesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// 이 도구는 실제 구현을 위해 Notion API 호출이 필요합니다.
		// 이 예제에서는 실제 API 호출 없이 샘플 응답만 반환합니다.

		limit := 5 // 기본값
		if limitVal, ok := request.Params.Arguments["limit"].(float64); ok {
			limit = int(limitVal)
		}

		// 샘플 데이터 반환
		result := fmt.Sprintf("Notion 데이터베이스 목록 (최대 %d개):\n\n", limit)
		result += "1. 프로젝트 관리\n"
		result += "   ID: 123456789abcdef\n"
		result += "   속성: 이름, 상태, 담당자, 마감일\n\n"
		result += "2. 제품 카탈로그\n"
		result += "   ID: abcdef123456789\n"
		result += "   속성: 제품명, 카테고리, 가격, 재고\n\n"
		result += "3. 팀원 디렉토리\n"
		result += "   ID: 9876543210abcde\n"
		result += "   속성: 이름, 직책, 부서, 이메일\n\n"

		return mcp.NewToolResultText(result), nil
	})

	return nil
}

// registerQueryDatabaseTool는 데이터베이스 쿼리 도구를 등록합니다.
func registerQueryDatabaseTool(server *mcpserver.MCPServer, client *NotionClient) error {
	// 데이터베이스 쿼리 도구 정의
	queryDatabaseTool := mcp.NewTool(ToolQueryDatabase,
		mcp.WithDescription("Notion 데이터베이스 쿼리"),
		mcp.WithString("database_id",
			mcp.Required(),
			mcp.Description("Notion 데이터베이스 ID"),
		),
		mcp.WithString("filter_json",
			mcp.Description("필터 JSON 문자열 (Notion API 형식)"),
		),
		mcp.WithString("sorts_json",
			mcp.Description("정렬 JSON 문자열 (Notion API 형식)"),
		),
		mcp.WithNumber("limit",
			mcp.Description("결과 제한 수"),
		),
	)

	// 데이터베이스 쿼리 도구 핸들러 등록
	server.AddTool(queryDatabaseTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// 이 도구는 실제 구현을 위해 Notion API 호출이 필요합니다.
		// 이 예제에서는 실제 API 호출 없이 샘플 응답만 반환합니다.

		databaseID, ok := request.Params.Arguments["database_id"].(string)
		if !ok || databaseID == "" {
			return mcp.NewToolResultError("유효한 데이터베이스 ID가 필요합니다"), nil
		}

		filterJSON, _ := request.Params.Arguments["filter_json"].(string)
		sortsJSON, _ := request.Params.Arguments["sorts_json"].(string)

		limit := 10 // 기본값
		if limitVal, ok := request.Params.Arguments["limit"].(float64); ok {
			limit = int(limitVal)
		}

		// 요청 파라미터 요약
		summary := fmt.Sprintf("데이터베이스 ID: %s\n", databaseID)
		if filterJSON != "" {
			summary += fmt.Sprintf("필터: %s\n", filterJSON)
		}
		if sortsJSON != "" {
			summary += fmt.Sprintf("정렬: %s\n", sortsJSON)
		}
		summary += fmt.Sprintf("제한: %d개\n\n", limit)

		// 샘플 데이터 반환
		result := fmt.Sprintf("Notion 데이터베이스 쿼리 결과:\n\n")
		result += summary
		result += "결과 항목:\n\n"
		result += "1. 프로젝트 관리 시스템 개발\n"
		result += "   상태: 진행 중\n"
		result += "   담당자: 홍길동\n"
		result += "   마감일: 2024-06-30\n\n"
		result += "2. 모바일 앱 로그인 기능 개선\n"
		result += "   상태: 계획됨\n"
		result += "   담당자: 김철수\n"
		result += "   마감일: 2024-07-15\n\n"
		result += "3. 데이터 분석 대시보드 구축\n"
		result += "   상태: 완료됨\n"
		result += "   담당자: 이영희\n"
		result += "   마감일: 2024-05-20\n\n"

		return mcp.NewToolResultText(result), nil
	})

	return nil
}

// min은 두 정수 중 작은 값을 반환합니다.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// generateFakeID는 테스트용 가짜 ID를 생성합니다.
func generateFakeID() string {
	return "abcdef1234567890"
}
