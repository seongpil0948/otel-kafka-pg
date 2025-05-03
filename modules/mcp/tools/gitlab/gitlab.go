package gitlab

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/xanzy/go-gitlab"
)

const (
	ToolCodeRecommend      = "gitlab_code_recommend"
	ToolSearchRepositories = "gitlab_search_repositories"
	ToolGetRepository      = "gitlab_get_repository"
	ToolListMergeRequests  = "gitlab_list_merge_requests"
	ToolGetMergeRequest    = "gitlab_get_merge_request"
)

// isToolEnabledFunc는 도구가 활성화되어 있는지 확인하는 함수 타입입니다.
type isToolEnabledFunc func(string) bool

// RegisterTools는 GitLab 관련 도구들을 서버에 등록합니다.
func RegisterTools(
	server *mcpserver.MCPServer,
	token string,
	baseURL string,
	isToolEnabled isToolEnabledFunc,
) error {
	// GitLab 클라이언트 생성
	var client *gitlab.Client
	var err error

	if baseURL != "" {
		// 사용자 지정 GitLab 인스턴스 URL 사용
		client, err = gitlab.NewClient(token, gitlab.WithBaseURL(baseURL))
	} else {
		// 기본 GitLab.com URL 사용
		client, err = gitlab.NewClient(token)
	}

	if err != nil {
		return fmt.Errorf("GitLab 클라이언트 생성 실패: %w", err)
	}

	// 코드 추천 도구 등록
	if isToolEnabled(ToolCodeRecommend) {
		if err := registerCodeRecommendTool(server, client); err != nil {
			return err
		}
	}

	// 저장소 검색 도구 등록
	if isToolEnabled(ToolSearchRepositories) {
		if err := registerSearchRepositoriesTool(server, client); err != nil {
			return err
		}
	}

	// 저장소 조회 도구 등록
	if isToolEnabled(ToolGetRepository) {
		if err := registerGetRepositoryTool(server, client); err != nil {
			return err
		}
	}

	// 머지 리퀘스트 목록 조회 도구 등록
	if isToolEnabled(ToolListMergeRequests) {
		if err := registerListMergeRequestsTool(server, client); err != nil {
			return err
		}
	}

	// 머지 리퀘스트 상세 조회 도구 등록
	if isToolEnabled(ToolGetMergeRequest) {
		if err := registerGetMergeRequestTool(server, client); err != nil {
			return err
		}
	}

	return nil
}

// registerCodeRecommendTool는 코드 컨텍스트 기반 추천 도구를 등록합니다.
func registerCodeRecommendTool(server *mcpserver.MCPServer, client *gitlab.Client) error {
	// 코드 추천 도구 정의
	codeRecommendTool := mcp.NewTool(ToolCodeRecommend,
		mcp.WithDescription("코드 컨텍스트 기반 변수명, 쿼리 등 추천"),
		mcp.WithString("context",
			mcp.Required(),
			mcp.Description("현재 코드 컨텍스트"),
		),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("GitLab 프로젝트 ID 또는 경로"),
		),
		mcp.WithString("recommendation_type",
			mcp.Required(),
			mcp.Description("추천 유형 (variable, query 등)"),
			mcp.Enum("variable", "query", "function"),
		),
	)

	// 코드 추천 도구 핸들러 등록
	server.AddTool(codeRecommendTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// 요청 파라미터 파싱
		codeContext, ok := request.Params.Arguments["context"].(string)
		if !ok || codeContext == "" {
			return mcp.NewToolResultError("유효한 코드 컨텍스트가 필요합니다"), nil
		}

		projectID, ok := request.Params.Arguments["project_id"].(string)
		if !ok || projectID == "" {
			return mcp.NewToolResultError("유효한 프로젝트 ID가 필요합니다"), nil
		}

		recommendationType, ok := request.Params.Arguments["recommendation_type"].(string)
		if !ok || recommendationType == "" {
			return mcp.NewToolResultError("유효한 추천 유형이 필요합니다"), nil
		}

		// 프로젝트 조회 및 코드 분석 로직
		project, _, err := client.Projects.GetProject(projectID, &gitlab.GetProjectOptions{})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("프로젝트 조회 실패: %v", err)), nil
		}

		// 이 예시에서는 실제 코드 분석 대신 간단한 응답 반환
		// 실제 구현에서는 프로젝트의 코드를 분석하고 패턴을 찾아 추천

		// 추천 유형에 따른 응답 생성
		var recommendation string
		switch recommendationType {
		case "variable":
			recommendation = "코드 컨텍스트와 프로젝트 패턴에 기반한 변수명 추천: userProfile, customerData"
		case "query":
			recommendation = "프로젝트 패턴에 기반한 쿼리 구조 추천: SELECT u.id, u.name FROM users u WHERE u.status = 'active'"
		case "function":
			recommendation = "패턴에 기반한 함수명 추천: processUserData, validateCustomerInput"
		default:
			return mcp.NewToolResultError("지원하지 않는 추천 유형입니다"), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("프로젝트: %s\n\n%s", project.Name, recommendation)), nil
	})

	return nil
}

// registerSearchRepositoriesTool는 저장소 검색 도구를 등록합니다.
func registerSearchRepositoriesTool(server *mcpserver.MCPServer, client *gitlab.Client) error {
	// 저장소 검색 도구 정의
	searchReposTool := mcp.NewTool(ToolSearchRepositories,
		mcp.WithDescription("GitLab 저장소 검색"),
		mcp.WithString("search",
			mcp.Required(),
			mcp.Description("검색어"),
		),
		mcp.WithNumber("limit",
			mcp.Description("결과 제한 수"),
		),
	)

	// 저장소 검색 도구 핸들러 등록
	server.AddTool(searchReposTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// 요청 파라미터 파싱
		search, ok := request.Params.Arguments["search"].(string)
		if !ok || search == "" {
			return mcp.NewToolResultError("유효한 검색어가 필요합니다"), nil
		}

		// 결과 제한 설정
		limit := 10 // 기본값
		if limitVal, ok := request.Params.Arguments["limit"].(float64); ok {
			limit = int(limitVal)
		}

		// GitLab 프로젝트 검색
		opts := &gitlab.ListProjectsOptions{
			Search: gitlab.String(search),
			ListOptions: gitlab.ListOptions{
				PerPage: limit,
				Page:    1,
			},
		}

		projects, _, err := client.Projects.ListProjects(opts)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("저장소 검색 실패: %v", err)), nil
		}

		// 검색 결과가 없는 경우
		if len(projects) == 0 {
			return mcp.NewToolResultText("검색 결과가 없습니다"), nil
		}

		// 검색 결과 포맷팅
		result := fmt.Sprintf("GitLab 저장소 검색 결과 (%d개):\n\n", len(projects))

		for i, project := range projects {
			result += fmt.Sprintf("%d. %s (ID: %d)\n", i+1, project.NameWithNamespace, project.ID)
			result += fmt.Sprintf("   설명: %s\n", project.Description)
			result += fmt.Sprintf("   URL: %s\n", project.WebURL)
			result += fmt.Sprintf("   기본 브랜치: %s\n", project.DefaultBranch)
			result += "\n"
		}

		return mcp.NewToolResultText(result), nil
	})

	return nil
}

// registerGetRepositoryTool는 저장소 조회 도구를 등록합니다.
func registerGetRepositoryTool(server *mcpserver.MCPServer, client *gitlab.Client) error {
	// 저장소 조회 도구 정의
	getRepoTool := mcp.NewTool(ToolGetRepository,
		mcp.WithDescription("GitLab 저장소 상세 정보 조회"),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("GitLab 프로젝트 ID 또는 경로"),
		),
	)

	// 저장소 조회 도구 핸들러 등록
	server.AddTool(getRepoTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// 요청 파라미터 파싱
		projectID, ok := request.Params.Arguments["project_id"].(string)
		if !ok || projectID == "" {
			return mcp.NewToolResultError("유효한 프로젝트 ID가 필요합니다"), nil
		}

		// GitLab 프로젝트 조회
		project, _, err := client.Projects.GetProject(projectID, &gitlab.GetProjectOptions{})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("저장소 조회 실패: %v", err)), nil
		}

		// 프로젝트 정보 포맷팅
		result := fmt.Sprintf("GitLab 저장소 정보:\n\n")
		result += fmt.Sprintf("이름: %s\n", project.NameWithNamespace)
		result += fmt.Sprintf("ID: %d\n", project.ID)
		result += fmt.Sprintf("설명: %s\n", project.Description)
		result += fmt.Sprintf("URL: %s\n", project.WebURL)
		result += fmt.Sprintf("기본 브랜치: %s\n", project.DefaultBranch)
		result += fmt.Sprintf("가시성: %s\n", project.Visibility)
		result += fmt.Sprintf("생성일: %s\n", project.CreatedAt.Format("2006-01-02 15:04:05"))

		if project.ForksCount > 0 {
			result += fmt.Sprintf("포크 수: %d\n", project.ForksCount)
		}

		if project.StarCount > 0 {
			result += fmt.Sprintf("스타 수: %d\n", project.StarCount)
		}

		return mcp.NewToolResultText(result), nil
	})

	return nil
}

// registerListMergeRequestsTool는 머지 리퀘스트 목록 조회 도구를 등록합니다.
func registerListMergeRequestsTool(server *mcpserver.MCPServer, client *gitlab.Client) error {
	// 머지 리퀘스트 목록 조회 도구 정의
	listMRTool := mcp.NewTool(ToolListMergeRequests,
		mcp.WithDescription("GitLab 프로젝트의 머지 리퀘스트 목록 조회"),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("GitLab 프로젝트 ID 또는 경로"),
		),
		mcp.WithString("state",
			mcp.Description("머지 리퀘스트 상태 (opened, closed, merged, all)"),
			mcp.Enum("opened", "closed", "merged", "all"),
		),
		mcp.WithNumber("limit",
			mcp.Description("결과 제한 수"),
		),
	)

	// 머지 리퀘스트 목록 조회 도구 핸들러 등록
	server.AddTool(listMRTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// 요청 파라미터 파싱
		projectID, ok := request.Params.Arguments["project_id"].(string)
		if !ok || projectID == "" {
			return mcp.NewToolResultError("유효한 프로젝트 ID가 필요합니다"), nil
		}

		// 상태 파라미터
		state := "opened" // 기본값
		if stateVal, ok := request.Params.Arguments["state"].(string); ok && stateVal != "" {
			state = stateVal
		}

		// 결과 제한 설정
		limit := 10 // 기본값
		if limitVal, ok := request.Params.Arguments["limit"].(float64); ok {
			limit = int(limitVal)
		}

		// GitLab 머지 리퀘스트 조회
		opts := &gitlab.ListProjectMergeRequestsOptions{
			ListOptions: gitlab.ListOptions{
				PerPage: limit,
				Page:    1,
			},
		}

		// 상태 필터 설정
		if state != "all" {
			opts.State = gitlab.String(state)
		}

		mrs, _, err := client.MergeRequests.ListProjectMergeRequests(projectID, opts)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("머지 리퀘스트 조회 실패: %v", err)), nil
		}

		// 조회 결과가 없는 경우
		if len(mrs) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("프로젝트 '%s'에서 '%s' 상태의 머지 리퀘스트가 없습니다", projectID, state)), nil
		}

		// 결과 포맷팅
		result := fmt.Sprintf("%s 상태의 머지 리퀘스트 목록 (%d개):\n\n", state, len(mrs))

		for i, mr := range mrs {
			result += fmt.Sprintf("%d. !%d %s\n", i+1, mr.IID, mr.Title)
			result += fmt.Sprintf("   작성자: %s\n", mr.Author.Username)
			result += fmt.Sprintf("   상태: %s\n", mr.State)
			result += fmt.Sprintf("   생성일: %s\n", mr.CreatedAt.Format("2006-01-02"))
			result += fmt.Sprintf("   URL: %s\n", mr.WebURL)
			result += "\n"
		}

		return mcp.NewToolResultText(result), nil
	})

	return nil
}

// registerGetMergeRequestTool는 머지 리퀘스트 상세 조회 도구를 등록합니다.
func registerGetMergeRequestTool(server *mcpserver.MCPServer, client *gitlab.Client) error {
	// 머지 리퀘스트 상세 조회 도구 정의
	getMRTool := mcp.NewTool(ToolGetMergeRequest,
		mcp.WithDescription("GitLab 머지 리퀘스트 상세 정보 조회"),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("GitLab 프로젝트 ID 또는 경로"),
		),
		mcp.WithNumber("merge_request_iid",
			mcp.Required(),
			mcp.Description("머지 리퀘스트 IID(내부 ID)"),
		),
	)

	// 머지 리퀘스트 상세 조회 도구 핸들러 등록
	server.AddTool(getMRTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// 요청 파라미터 파싱
		projectID, ok := request.Params.Arguments["project_id"].(string)
		if !ok || projectID == "" {
			return mcp.NewToolResultError("유효한 프로젝트 ID가 필요합니다"), nil
		}

		mrIIDFloat, ok := request.Params.Arguments["merge_request_iid"].(float64)
		if !ok {
			return mcp.NewToolResultError("유효한 머지 리퀘스트 IID가 필요합니다"), nil
		}
		mrIID := int(mrIIDFloat)

		// GitLab 머지 리퀘스트 조회
		mr, _, err := client.MergeRequests.GetMergeRequest(projectID, mrIID, &gitlab.GetMergeRequestsOptions{})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("머지 리퀘스트 조회 실패: %v", err)), nil
		}

		// 머지 리퀘스트 변경 사항 조회
		changes, _, err := client.MergeRequests.GetMergeRequestChanges(projectID, mrIID, &gitlab.GetMergeRequestChangesOptions{})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("머지 리퀘스트 변경 사항 조회 실패: %v", err)), nil
		}

		// 머지 리퀘스트 정보 포맷팅
		result := fmt.Sprintf("머지 리퀘스트 상세 정보:\n\n")
		result += fmt.Sprintf("제목: %s\n", mr.Title)
		result += fmt.Sprintf("설명: %s\n\n", mr.Description)
		result += fmt.Sprintf("작성자: %s (@%s)\n", mr.Author.Name, mr.Author.Username)
		result += fmt.Sprintf("상태: %s\n", mr.State)
		result += fmt.Sprintf("소스 브랜치: %s\n", mr.SourceBranch)
		result += fmt.Sprintf("대상 브랜치: %s\n", mr.TargetBranch)
		result += fmt.Sprintf("생성일: %s\n", mr.CreatedAt.Format("2006-01-02 15:04:05"))

		if mr.MergedAt != nil {
			result += fmt.Sprintf("머지일: %s\n", mr.MergedAt.Format("2006-01-02 15:04:05"))
		}

		result += fmt.Sprintf("URL: %s\n\n", mr.WebURL)

		// 변경 파일 목록
		result += fmt.Sprintf("변경된 파일 (%d개):\n", len(changes.Changes))
		for i, change := range changes.Changes {
			result += fmt.Sprintf("%d. %s (%s)\n", i+1, change.NewPath, change.NewFile)
		}

		return mcp.NewToolResultText(result), nil
	})

	return nil
}
