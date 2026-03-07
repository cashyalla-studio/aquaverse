# AquaVerse Makefile

.PHONY: server web app infra migrate dev-up dev-down build-all \
        local-up local-down local-build local-logs local-ps local-clean \
        app-run-android app-run-ios app-run-chrome app-run-device app-codegen \
        app-deploy-ios app-deploy-android app-deploy-all \
        monitoring-up monitoring-down

# ── 서버 ──────────────────────────────────────────────
server-run:
	cd server && go run ./cmd/server/...

server-build:
	cd server && CGO_ENABLED=0 GOOS=linux go build -o bin/aquaverse-api ./cmd/server/...

server-test:
	cd server && go test ./... -v -race

server-lint:
	cd server && golangci-lint run ./...

server-tidy:
	cd server && go mod tidy

# ── 데이터베이스 마이그레이션 ──────────────────────────
migrate-up:
	migrate -path server/migrations -database "$(AV_DATABASE_DSN)" up

migrate-down:
	migrate -path server/migrations -database "$(AV_DATABASE_DSN)" down 1

migrate-status:
	migrate -path server/migrations -database "$(AV_DATABASE_DSN)" version

# ── 웹 ────────────────────────────────────────────────
web-dev:
	cd web && npm run dev

web-build:
	cd web && npm run build

web-install:
	cd web && npm install

# ── Flutter ───────────────────────────────────────────
app-run:
	cd app && flutter run

app-build-apk:
	cd app && flutter build apk --release

app-build-ios:
	cd app && flutter build ios --release

app-l10n:
	cd app && flutter gen-l10n

app-get:
	cd app && flutter pub get

# ── Docker ────────────────────────────────────────────
dev-up:
	docker compose -f infra/docker-compose.dev.yml up -d

dev-down:
	docker compose -f infra/docker-compose.dev.yml down

stack-deploy:
	docker stack deploy -c infra/docker-stack.yml aquaverse

stack-remove:
	docker stack rm aquaverse

stack-ps:
	docker stack ps aquaverse

# ── 빌드 & 배포 ───────────────────────────────────────
build-server-image:
	docker build -t $(REGISTRY)/aquaverse/api:$(VERSION) -f server/Dockerfile server/

build-web-image:
	docker build -t $(REGISTRY)/aquaverse/web:$(VERSION) -f web/Dockerfile web/

push-images: build-server-image build-web-image
	docker push $(REGISTRY)/aquaverse/api:$(VERSION)
	docker push $(REGISTRY)/aquaverse/web:$(VERSION)

# ── 로컬 통합 테스트 (Docker Compose) ─────────────────
## 최초 1회 또는 코드 변경 시: make local-up
local-up:
	docker compose -f docker-compose.local.yml up --build -d
	@echo ""
	@echo "🚀 AquaVerse 로컬 실행 중"
	@echo "   웹:          http://localhost:3000"
	@echo "   API:         http://localhost:8080"
	@echo "   API 문서:    http://localhost:8080/health"
	@echo "   MinIO 콘솔:  http://localhost:9001  (minioadmin / minioadmin123)"
	@echo "   PostgreSQL:  localhost:5432  (aquaverse / devpassword)"
	@echo ""
	@echo "로그 확인: make local-logs"

## 서비스 중단 (볼륨 유지)
local-down:
	docker compose -f docker-compose.local.yml down

## 빌드만 (시작 안 함)
local-build:
	docker compose -f docker-compose.local.yml build --no-cache

## 로그 스트리밍
local-logs:
	docker compose -f docker-compose.local.yml logs -f --tail=100

## API 서버 로그만
local-logs-api:
	docker compose -f docker-compose.local.yml logs -f --tail=100 api

## 컨테이너 상태 확인
local-ps:
	docker compose -f docker-compose.local.yml ps

## 완전 초기화 (볼륨 포함 삭제)
local-clean:
	docker compose -f docker-compose.local.yml down -v --remove-orphans
	@echo "✅ 볼륨 및 컨테이너 모두 삭제됨"

## API만 재시작 (코드 변경 시)
local-restart-api:
	docker compose -f docker-compose.local.yml build api
	docker compose -f docker-compose.local.yml up -d api

## 웹만 재시작 (프론트엔드 코드 변경 시)
local-restart-web:
	docker compose -f docker-compose.local.yml build web
	docker compose -f docker-compose.local.yml up -d web

# ── Flutter 로컬 실행 ──────────────────────────────────
## 코드 생성 (Riverpod, l10n 등)
app-codegen:
	cd app && flutter pub get
	cd app && flutter gen-l10n
	cd app && flutter pub run build_runner build --delete-conflicting-outputs

## Android 에뮬레이터 (API: 10.0.2.2:8080)
app-run-android:
	cd app && flutter run -d android \
	  --dart-define="API_URL=http://10.0.2.2:8080" \
	  --dart-define="ENV=local"

## iOS 시뮬레이터 (API: localhost:8080)
app-run-ios:
	cd app && flutter run -d ios \
	  --dart-define="API_URL=http://localhost:8080" \
	  --dart-define="ENV=local"

## Chrome 브라우저
app-run-chrome:
	cd app && flutter run -d chrome \
	  --dart-define="API_URL=http://localhost:8080" \
	  --dart-define="ENV=local"

## Windows 데스크탑
app-run-windows:
	cd app && flutter run -d windows \
	  --dart-define="API_URL=http://localhost:8080" \
	  --dart-define="ENV=local"

## 실물 기기 (HOST_IP 수동 지정 필요: make app-run-device HOST_IP=192.168.1.x)
app-run-device:
	cd app && flutter run \
	  --dart-define="API_URL=http://$(HOST_IP):8080" \
	  --dart-define="ENV=local"

## 자동 플랫폼 감지 실행 스크립트
app-run:
	bash scripts/flutter-run-local.sh $(PLATFORM)

## Monitoring
monitoring-up: ## Prometheus + Grafana 시작
	docker compose -f docker-compose.local.yml up -d prometheus grafana

monitoring-down: ## Prometheus + Grafana 중지
	docker compose -f docker-compose.local.yml stop prometheus grafana

# ── App Store 배포 ──────────────────────────────────
.PHONY: app-deploy-ios app-deploy-android app-deploy-all

## iOS TestFlight 배포 (Fastlane)
app-deploy-ios:
	cd app && bundle exec fastlane ios beta

## Android Play Store Internal 배포
app-deploy-android:
	cd app && bundle exec fastlane android beta

## 전체 플랫폼 배포
app-deploy-all: app-deploy-ios app-deploy-android

# ── 전체 ─────────────────────────────────────────────
build-all: server-build web-build

setup: server-tidy web-install app-get
	@echo "✓ AquaVerse setup complete"
