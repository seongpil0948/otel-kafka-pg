package server

// Config는 MCP 서버 설정 구조체입니다.
type Config struct {
	// 외부 서비스 API 토큰
	GitLabToken string
	GitLabURL   string
	NotionToken string
	FigmaToken  string
	
	// 서버 설정
	Port        int
	Host        string
	LogEnabled  bool
	
	// 기능 설정
	EnabledTools []string
}

// DefaultConfig는 기본 MCP 서버 설정을 반환합니다.
func DefaultConfig() *Config {
	return &Config{
		Port:        8090,
		Host:        "localhost",
		LogEnabled:  true,
		EnabledTools: []string{},
	}
}
