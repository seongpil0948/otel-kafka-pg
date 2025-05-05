#!/bin/bash
set -e

API_URL=${1:-"http://localhost:8080/api/telemetry"}
ENDPOINT=${2:-"logs"}
PARAMS=${3:-"limit=10"}

FULL_URL="${API_URL}/${ENDPOINT}?${PARAMS}"

echo "==== REST API 캐싱 테스트 ===="
echo "테스트 URL: $FULL_URL"

# 첫 번째 요청
echo -e "\n---- 첫 번째 요청 (캐시 미스 예상) ----"
time curl -s "$FULL_URL" > /dev/null

# 두 번째 요청
echo -e "\n---- 두 번째 요청 (캐시 히트 예상) ----"
time curl -s "$FULL_URL" > /dev/null

# 세 번째 요청
echo -e "\n---- 세 번째 요청 (캐시 히트 예상) ----"
time curl -s "$FULL_URL" > /dev/null

echo -e "\n캐싱이 제대로 작동한다면 두 번째와 세 번째 요청이 첫 번째보다 훨씬 빨라야 합니다."
echo "API 서버 로그에서 'cache hit'과 'cache miss' 메시지를 확인하세요."
echo -e "\n==== 테스트 완료 ===="