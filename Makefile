# AquaVerse Makefile

.PHONY: server web app infra migrate dev-up dev-down build-all

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

# ── 전체 ─────────────────────────────────────────────
build-all: server-build web-build

setup: server-tidy web-install app-get
	@echo "✓ AquaVerse setup complete"
