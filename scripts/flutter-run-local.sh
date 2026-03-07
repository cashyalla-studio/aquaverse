#!/usr/bin/env bash
# AquaVerse Flutter 로컬 실행 스크립트
# 사용법: bash scripts/flutter-run-local.sh [android|ios|chrome|windows]
#
# 전제조건: docker compose -f docker-compose.local.yml up -d api 가 먼저 실행돼야 함

set -e

PLATFORM=${1:-android}
APP_DIR="$(cd "$(dirname "$0")/../app" && pwd)"

# 플랫폼별 API URL 설정
# - Android 에뮬레이터: 10.0.2.2 (호스트 루프백)
# - iOS 시뮬레이터: localhost
# - 실물 기기: 호스트 머신의 실제 IP (자동 감지)
# - Chrome/Web: localhost
# - Windows: localhost

case "$PLATFORM" in
  android)
    # Android 에뮬레이터는 10.0.2.2가 호스트 localhost에 매핑
    API_URL="http://10.0.2.2:8080"
    ;;
  ios)
    API_URL="http://localhost:8080"
    ;;
  chrome|web)
    API_URL="http://localhost:8080"
    PLATFORM="chrome"
    ;;
  windows)
    API_URL="http://localhost:8080"
    ;;
  device)
    # 실물 기기용: 호스트 IP 자동 감지
    if command -v ip &>/dev/null; then
      HOST_IP=$(ip route get 1.1.1.1 | awk '{print $7; exit}')
    elif command -v ipconfig &>/dev/null; then
      HOST_IP=$(ipconfig | grep -A4 "Wireless\|Ethernet" | grep "IPv4" | head -1 | awk '{print $NF}' | tr -d '\r')
    else
      HOST_IP="192.168.1.100"  # 수동 설정 필요
    fi
    API_URL="http://${HOST_IP}:8080"
    echo "📱 실물 기기용 API URL: $API_URL"
    echo "   (자동 감지 실패 시 스크립트에서 HOST_IP를 수동 지정하세요)"
    PLATFORM="--no-specify"
    ;;
  *)
    echo "사용법: $0 [android|ios|chrome|windows|device]"
    exit 1
    ;;
esac

echo "🚀 AquaVerse Flutter 실행"
echo "   플랫폼: $PLATFORM"
echo "   API URL: $API_URL"
echo ""

# 의존성 확인
cd "$APP_DIR"

if ! flutter doctor -v &>/dev/null; then
  echo "❌ Flutter SDK가 없거나 경로가 잘못됐습니다"
  exit 1
fi

# 패키지 설치
echo "📦 패키지 설치 중..."
flutter pub get

# l10n 생성
echo "🌐 다국어 파일 생성 중..."
flutter gen-l10n 2>/dev/null || true

# 코드 생성 (Riverpod 등)
if flutter pub run build_runner build --delete-conflicting-outputs 2>/dev/null; then
  echo "✅ 코드 생성 완료"
fi

# API 서버 연결 확인
API_CHECK_URL="${API_URL/10.0.2.2/localhost}"  # 에뮬레이터 주소를 로컬 주소로 변환해서 체크
if curl -sf "${API_CHECK_URL}/health" &>/dev/null; then
  echo "✅ API 서버 연결 확인됨 (${API_CHECK_URL})"
else
  echo "⚠️  API 서버에 연결할 수 없습니다 (${API_CHECK_URL})"
  echo "   먼저 실행하세요: docker compose -f docker-compose.local.yml up -d api"
  echo "   계속 진행하시겠습니까? [y/N]"
  read -r answer
  [[ "$answer" =~ ^[Yy]$ ]] || exit 1
fi

echo ""
echo "▶  Flutter 앱 실행 중..."

# 플랫폼별 실행
if [[ "$PLATFORM" == "--no-specify" ]]; then
  flutter run \
    --dart-define="API_URL=${API_URL}" \
    --dart-define="ENV=local"
else
  flutter run \
    -d "$PLATFORM" \
    --dart-define="API_URL=${API_URL}" \
    --dart-define="ENV=local"
fi
