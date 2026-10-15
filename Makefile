DEV_IMAGE := gen-asyncapi-dev
UID := $(shell id -u)
GID := $(shell id -g)
MIN_TEST_COVERAGE := 90

# Every Go and tooling command runs inside the dev image, so the toolchain
# (Go 1.26.9), the linter and the security scanners are identical in a local
# shell and in CI. The workspace is bind-mounted; caches go to /tmp so the
# non-root run needs no writable GOPATH. CGO stays on for the race detector.
DEV_RUN := docker run --rm --init \
	--user $(UID):$(GID) \
	--cap-drop ALL \
	--security-opt no-new-privileges:true \
	-e HOME=/tmp \
	-e GOPATH=/tmp/go \
	-e GOCACHE=/tmp/go-build \
	-e CGO_ENABLED=1 \
	-e MIN_TEST_COVERAGE=$(MIN_TEST_COVERAGE) \
	-v $(CURDIR):/work \
	-w /work \
	$(DEV_IMAGE)

.PHONY: all dev-image dep format generate lint lint-fix shell test test-unit \
	test-coverage test-ts sec pkg-lock pkg-add pkg-update pkg-upgrade \
	pkg-remove help

all: dep lint test ## Run dep, lint and test

dev-image: ## Build the sandboxed development image
	@docker build -f Dockerfile.dev -t $(DEV_IMAGE) .

dep: dev-image ## Get project dependencies (go mod tidy + vendor)
	@$(DEV_RUN) sh -ceu 'go mod tidy && go mod vendor'

format: dev-image ## Format Go source
	@$(DEV_RUN) sh -ceu 'find . -path ./vendor -prune -o -path ./.research_files -prune -o -type f -name "*.go" -exec gofmt -w {} +'
	@$(DEV_RUN) shfmt -w scripts/*.sh

generate: dev-image ## Run all code generation
	@$(DEV_RUN) go generate ./...

lint: dev-image ## Lint all Golang files
	@$(DEV_RUN) bash scripts/lint.sh

lint-fix: dev-image ## Lint all Golang files and fix
	@$(DEV_RUN) sh -ceu 'go fix ./...; go tool golangci-lint run --fix --timeout=30m0s ./...'

test: dev-image ## Run all tests
	@$(DEV_RUN) go test -race -count=1 -timeout=10m ./...
	@$(DEV_RUN) bash scripts/test-ts.sh

test-ts: dev-image ## Type-check and test the generated TypeScript
	@$(DEV_RUN) bash scripts/test-ts.sh

test-unit: test ## Run the unit test suite

shell: dev-image ## Open a shell in the development image
	@$(DEV_RUN) sh

test-coverage: dev-image ## Run tests with coverage check. Fails if coverage is below the threshold.
	@$(DEV_RUN) bash scripts/test-coverage.sh
	@$(DEV_RUN) bash scripts/test-ts.sh

# The TypeScript check tools are pnpm dev dependencies, installed into the dev
# image from pnpm-lock.yaml. These targets only rewrite package.json and the
# lockfile; run make dev-image afterwards. pnpm-workspace.yaml holds the
# seven-day release-age gate, and no install runs package scripts.
pkg-lock: dev-image ## Refresh pnpm-lock.yaml under the existing pins and age gate
	@$(DEV_RUN) pnpm install --lockfile-only --ignore-scripts

pkg-add: dev-image ## Add an exact TypeScript dev dependency: make pkg-add PKG=name@version
	@test -n "$(PKG)" || { echo "PKG is required" >&2; exit 1; }
	@$(DEV_RUN) pnpm add --save-dev --save-exact --lockfile-only --ignore-scripts $(PKG)

pkg-update: dev-image ## Update one TypeScript dev dependency: make pkg-update PKG=name
	@test -n "$(PKG)" || { echo "PKG is required" >&2; exit 1; }
	@$(DEV_RUN) pnpm update --latest --lockfile-only --ignore-scripts $(PKG)

pkg-upgrade: dev-image ## Upgrade every TypeScript dev dependency to its newest gated release
	@$(DEV_RUN) pnpm update --latest --lockfile-only --ignore-scripts

pkg-remove: dev-image ## Remove a TypeScript dev dependency: make pkg-remove PKG=name
	@test -n "$(PKG)" || { echo "PKG is required" >&2; exit 1; }
	@$(DEV_RUN) pnpm remove --lockfile-only --ignore-scripts $(PKG)

sec: dev-image ## Security scan (gitleaks, govulncheck, pnpm audit, semgrep) merged into sec.sarif; gates on findings
	@$(DEV_RUN) bash scripts/sec.sh

help: ## Display this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
