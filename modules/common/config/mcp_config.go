package config

// MCPConfig는 MCP 서버 관련 설정입니다.
type MCPConfig struct {
	// 외부 서비스 API 토큰
	GitLabToken string `json:"gitlabToken" mapstructure:"gitlab_token"`
	GitLabURL   string `json:"gitlabUrl" mapstructure:"gitlab_url"`
	NotionToken string `json:"notionToken" mapstructure:"notion_token"`
	FigmaToken  string `json:"figmaToken" mapstructure:"figma_token"`
	
	// 기능 설정
	EnabledTools []string `json:"enabledTools" mapstructure:"enabled_tools"`
	
	// 서버 설정
	Port        int  `json:"port" mapstructure:"port"`
	LogEnabled  bool `json:"logEnabled" mapstructure:"log_enabled"`
}

// DefaultMCPConfig는 기본 MCP 설정을 반환합니다.
func DefaultMCPConfig() *MCPConfig {
	return &MCPConfig{
		Port:        8090,
		LogEnabled:  false,
		EnabledTools: []string{},
	}
}
