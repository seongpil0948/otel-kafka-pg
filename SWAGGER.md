# Swagger 문서 사용 가이드

이 프로젝트는 [Swagger](https://swagger.io/)를 사용하여 API 문서를 자동으로 생성합니다. Swagger는 API를 설계, 빌드, 문서화 및 테스트하는 데 도움이 되는 도구입니다.

## Swagger 문서 생성하기

API 문서를 생성하려면 다음 명령어를 실행하세요:

```bash
make swagger
```

이 명령어는 코드의 주석을 분석하여 `docs` 디렉토리에 Swagger 문서를 생성합니다.

## Swagger 문서 서식 정리하기

Swagger 주석 서식을 정리하려면 다음 명령어를 실행하세요:

```bash
make swagger-fmt
```

## Swagger UI 접근하기

애플리케이션이 실행 중일 때, 다음 URL을 통해 Swagger UI에 접근할 수 있습니다:

```
http://localhost:8080/swagger/index.html
```

## Swagger 주석 작성 방법

### API 주석 예시

API 엔드포인트에 다음과 같은 주석을 추가하여 Swagger 문서에 포함시킬 수 있습니다:

```go
// GetTraceByID godoc
// @Summary      트레이스 ID로 상세 정보 조회
// @Description  특정 트레이스 ID에 대한 상세 정보를 조회합니다
// @Tags         traces
// @Accept       json
// @Produce      json
// @Param        traceId path string true "Trace ID"
// @Success      200 {object} dto.Response{data=dto.TraceDetailResponse}
// @Failure      400 {object} dto.Response
// @Failure      404 {object} dto.Response
// @Failure      500 {object} dto.Response
// @Router       /telemetry/traces/{traceId} [get]
func (c *TraceController) GetTraceByID(ctx *gin.Context) {
    // 함수 구현
}
```

### 주요 주석 태그

- `@Summary`: API 엔드포인트의 간단한 요약
- `@Description`: API 엔드포인트의 상세 설명
- `@Tags`: API 엔드포인트를 그룹화하는 태그
- `@Accept`: 허용되는 요청 콘텐츠 타입
- `@Produce`: 응답 콘텐츠 타입
- `@Param`: 파라미터 정의 (이름, 위치, 타입, 필수 여부, 설명)
- `@Success`: 성공 응답 정의
- `@Failure`: 실패 응답 정의
- `@Router`: API 경로 및 HTTP 메서드

## 추가 참고 자료

- [Swaggo 공식 문서](https://github.com/swaggo/swag)
- [Swagger 공식 웹사이트](https://swagger.io/)
