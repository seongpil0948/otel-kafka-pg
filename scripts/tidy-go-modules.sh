#!/bin/bash
set -e

# 스크립트 시작 메시지
echo "===== Go 모듈 정리 스크립트 시작 ====="
echo "이 스크립트는 프로젝트의 모든 Go 모듈을 찾아 'go mod tidy'를 실행합니다."

# 프로젝트 루트 디렉토리 (스크립트가 루트에 있다고 가정)
ROOT_DIR=$(pwd)
echo "프로젝트 루트 디렉토리: $ROOT_DIR"

# go.mod 파일들을 찾아서 각 디렉토리에서 go mod tidy 실행
echo "go.mod 파일 검색 중..."
find "$ROOT_DIR" -name "go.mod" | while read -r module_file; do
    module_dir=$(dirname "$module_file")
    module_name=$(basename "$module_dir")
    
    echo -e "\n------------------------------------"
    echo "모듈 정리 중: $module_name ($module_dir)"
    
    # 모듈 디렉토리로 이동
    cd "$module_dir"
    
    # go mod tidy 실행
    echo "go mod tidy 실행 중..."
    go mod tidy
    
    # 결과 확인
    if [ $? -eq 0 ]; then
        echo "✅ $module_name 모듈 정리 완료"
    else
        echo "❌ $module_name 모듈 정리 실패"
        exit 1
    fi
    
    # 루트 디렉토리로 돌아가기
    cd "$ROOT_DIR"
done

# go.work 파일이 있다면 go work sync 실행
if [ -f "$ROOT_DIR/go.work" ]; then
    echo -e "\n------------------------------------"
    echo "go.work 파일 발견: go work sync 실행 중..."
    go work sync
    if [ $? -eq 0 ]; then
        echo "✅ go work sync 완료"
    else
        echo "❌ go work sync 실패"
        exit 1
    fi
fi

echo -e "\n===== 모든 Go 모듈 정리 완료 ====="