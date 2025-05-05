#!/bin/bash
set -e

echo "==== Redis 연결 상태 확인 ===="

# Redis 설정 확인
REDIS_HOST=${REDIS_ADDRESS:-redis:6379}
REDIS_HOST_PARTS=(${REDIS_HOST//:/ })
REDIS_HOST_NAME=${REDIS_HOST_PARTS[0]}
REDIS_PORT=${REDIS_HOST_PARTS[1]:-6379}
REDIS_ENABLED=${REDIS_ENABLE_CACHE:-true}
REDIS_TTL=${REDIS_TTL:-3600}

echo "Redis 호스트: $REDIS_HOST_NAME"
echo "Redis 포트: $REDIS_PORT"
echo "캐싱 활성화: $REDIS_ENABLED"
echo "캐시 TTL: $REDIS_TTL 초"

# Redis CLI 명령 실행
if command -v redis-cli > /dev/null; then
    echo -e "\n---- Redis 서버 정보 ----"
    redis-cli -h $REDIS_HOST_NAME -p $REDIS_PORT info | grep -E 'redis_version|connected_clients|used_memory_human|total_connections_received'
    
    echo -e "\n---- Redis 키 통계 ----"
    redis-cli -h $REDIS_HOST_NAME -p $REDIS_PORT keys 'api:cache:*' | wc -l | awk '{print "캐시 키 개수:", $1}'
    
    # 캐시 키 샘플 보기
    echo -e "\n---- 캐시 키 샘플 (최대 5개) ----"
    redis-cli -h $REDIS_HOST_NAME -p $REDIS_PORT keys 'api:cache:*' | head -5
    
    # 캐시 TTL 확인
    SAMPLE_KEY=$(redis-cli -h $REDIS_HOST_NAME -p $REDIS_PORT keys 'api:cache:*' | head -1)
    if [ ! -z "$SAMPLE_KEY" ]; then
        echo -e "\n---- 샘플 키 TTL 확인 ----"
        TTL=$(redis-cli -h $REDIS_HOST_NAME -p $REDIS_PORT ttl "$SAMPLE_KEY")
        echo "키: $SAMPLE_KEY"
        echo "남은 TTL: $TTL 초"
    fi
else
    echo "redis-cli가 설치되지 않았습니다. Docker를 통해 접근하세요:"
    echo "docker exec -it telemetry-redis redis-cli"
fi

echo -e "\n==== Redis 확인 완료 ===="