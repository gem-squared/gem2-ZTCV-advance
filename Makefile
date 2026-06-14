.PHONY: dev stop clean logs build test fmt vet healthcheck

# ─── Default ─────────────────────────────────────────────────────────
help:
	@echo "ZTCV — Zero-Trust Call Verification Protocol"
	@echo ""
	@echo "Common targets:"
	@echo "  make dev          boot the whole stack (6 containers) via docker compose"
	@echo "  make stop         docker compose down"
	@echo "  make clean        docker compose down -v + rm -rf data/*.db"
	@echo "  make logs SVC=x   tail logs for one service (session-svc / identity-svc / ...)"
	@echo "  make build        local go build ./cmd/..."
	@echo "  make test         local go test ./..."
	@echo "  make fmt          gofmt -w ."
	@echo "  make vet          go vet ./..."
	@echo "  make healthcheck  curl /healthz on all 4 backend binaries"

# ─── Docker compose ──────────────────────────────────────────────────
dev:
	docker compose up --build -d
	@echo ""
	@echo "Stack starting in background. Tail logs with: make logs SVC=session-svc"
	@echo "Health: make healthcheck"

stop:
	docker compose down

clean:
	docker compose down -v
	rm -rf data/*.db data/*.db-journal data/*.db-wal data/*.db-shm 2>/dev/null || true

logs:
	@if [ -z "$(SVC)" ]; then \
		echo "usage: make logs SVC=<session-svc|identity-svc|risk-chain-svc|chain-adapter|chain-sim|frontend>"; \
		exit 1; \
	fi
	docker compose logs -f $(SVC)

# ─── Local Go ────────────────────────────────────────────────────────
build:
	go build ./cmd/...

test:
	go test ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

# ─── Smoke ───────────────────────────────────────────────────────────
healthcheck:
	@for port in 8001 8002 8003 8004; do \
		printf ":%s → " "$$port"; \
		curl -fsS http://localhost:$$port/healthz || echo "FAIL"; \
		echo; \
	done
