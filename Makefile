# custos — developer entry points. `make verify` is the definition of done.

EA_PATH ?= $(HOME)/Sites/phpinspectionsea
RULE    ?=
GO      ?= go
BIN     := bin/custos

STUBS_REPO ?= https://github.com/JetBrains/phpstorm-stubs

FIXCHECK ?= 1

.PHONY: architecture coverage fixcheck stubs build test vet fmt fmt-check lint bench fuzz extract rules-doc fixtures conformance cleanroom verify clean docs docs-dev

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

build:
	$(GO) build -trimpath -ldflags "-X main.version=$(VERSION)" -o $(BIN) ./cmd/custos

test:
	$(GO) test ./...

# Linters are pinned and run through `go run` (built once with the local Go,
# then cached); markdownlint comes from docs/package.json (`npm ci --prefix
# docs`). Set any of them empty to skip it, e.g. `make lint MDLINT=` offline.
GOLANGCI   ?= $(GO) run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0
ACTIONLINT ?= $(GO) run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12
ECCHECK    ?= $(GO) run github.com/editorconfig-checker/editorconfig-checker/v3/cmd/editorconfig-checker@v3.11.3
MDLINT     ?= docs/node_modules/.bin/markdownlint-cli2

# Rules: .golangci.yml, .editorconfig (+ .editorconfig-checker.json),
# .markdownlint-cli2.jsonc. See docs/contributing/index.md.
lint:
	@$(if $(GOLANGCI),$(GOLANGCI) run ./...,echo "golangci-lint skipped")
	@$(if $(ACTIONLINT),$(ACTIONLINT),echo "actionlint skipped")
	@$(if $(ECCHECK),$(ECCHECK),echo "editorconfig-checker skipped")
	@$(if $(MDLINT),if [ ! -x "$(MDLINT)" ]; then echo "markdownlint missing: run npm ci --prefix docs (or MDLINT= to skip)"; exit 1; fi; $(MDLINT),echo "markdownlint skipped")

# Apply the formatters (gofumpt, goimports) and markdownlint's fixes.
fmt:
	$(GOLANGCI) fmt
	@if [ -n "$(MDLINT)" ]; then $(MDLINT) --fix; fi

# Kept as shortcuts: both are part of `make lint`.
vet:
	$(GOLANGCI) run --enable-only govet ./...

fmt-check:
	$(GOLANGCI) fmt --diff

# Benchmarks on a real file need CUSTOS_BENCH_FILE (a large PHP source).
bench:
	$(GO) test -run '^$$' -bench . -benchmem ./...
	CUSTOS_PERF=1 $(GO) test -count=1 -run TestEditLatency -v ./internal/editor/lsp

# Short fuzz smoke over every Fuzz* target (parser/lexer once they exist).
fuzz:
	@for pkg in $$($(GO) list ./...); do \
		for f in $$($(GO) test -list '^Fuzz' $$pkg | grep '^Fuzz'); do \
			$(GO) test -run '^$$' -fuzz "^$$f$$" -fuzztime 30s $$pkg || exit 1; \
		done; \
	done

# Rebuild the embedded builtin symbol index from phpstorm-stubs (Apache-2.0).
stubs:
	@ls -d .cache/stubs-src/*phpstorm-stubs* >/dev/null 2>&1 || \
		(mkdir -p .cache/stubs-src && git clone --depth 1 $(STUBS_REPO) .cache/stubs-src/phpstorm-stubs)
	$(GO) run ./tools/genstubs

# Regenerate rule facts (committed) and the local-only EA case index (.cache/).
extract:
	$(GO) run ./tools/extract -ea "$(EA_PATH)"
	$(MAKE) rules-doc

rules-doc:
	$(GO) run ./tools/rulesdoc
	$(GO) run ./tools/genexplain
	$(GO) run ./tools/rulesref

# Documentation site (VitePress, Node 20+): build to docs/.vitepress/dist, or
# serve with live reload.
docs: rules-doc
	npm --prefix docs ci
	npm --prefix docs run build

docs-dev: rules-doc
	npm --prefix docs install
	npm --prefix docs run dev

# custos' own fixtures (CI gate).
fixtures:
	$(GO) test ./internal/testing/conformance -run 'TestOwnFixtures' -rule "$(RULE)"

# EA fixtures read from a local checkout; never copied into this repo.
conformance:
	EA_PATH="$(EA_PATH)" $(GO) test ./internal/testing/conformance -run 'TestEA' -v -rule "$(RULE)"

# Verbatim-text scan against the local EA checkout (skipped when absent).
cleanroom:
	$(GO) run ./tools/cleanroom -ea "$(EA_PATH)"

# Apply every quick-fix offered on the real-world corpus (CUSTOS_CORPUS, a list
# of directories) and require the result to still parse (FIXCHECK=php
# additionally runs php -l on samples).
fixcheck:
	CUSTOS_FIXCHECK=$(FIXCHECK) $(GO) test ./internal/inspection/catalogue -run TestFixesKeepCodeParsable -v -timeout 30m

architecture:
	$(GO) run ./tools/architecture

verify: architecture lint test fixtures coverage cleanroom

# Coverage gates (the EA run and local corpora do not count):
# - every statement of internal/inspection/rules is covered by own fixtures and the rule
#   packages' tests;
# - every statement of cmd/, internal/ and tools/ is covered by the whole test
#   suite. tools/covercheck exempts only the body of a one-statement
#   `func main()` in package main (each command's os.Exit(run(...)) wrapper),
#   found from the source, so nothing has to be listed or kept in sync.
# Bound coverage test processes: each binary instruments the complete module.
COVER_SKIP := TestEA|TestNoCrashOnCorpus|TestFixesKeepCodeParsable
coverage:
	@mkdir -p .cache
	go test -p 4 ./internal/testing/conformance ./internal/inspection/catalogue ./internal/inspection/rules/... -skip '$(COVER_SKIP)' \
		-coverpkg=./internal/inspection/rules/... -coverprofile=.cache/rules-cover.out >.cache/rules-cover.log 2>&1 \
		|| { grep -E -B2 -A20 '^(--- FAIL|FAIL|panic:)' .cache/rules-cover.log; exit 1; }
	@$(GO) run ./tools/covercheck -what internal/inspection/rules .cache/rules-cover.out
	go test -p 4 ./... -skip '$(COVER_SKIP)' -coverpkg=./cmd/...,./internal/...,./tools/... -coverprofile=.cache/all-cover.out >.cache/all-cover.log 2>&1 \
		|| { grep -E -B2 -A20 '^(--- FAIL|FAIL|panic:)' .cache/all-cover.log; exit 1; }
	@$(GO) run ./tools/covercheck -what "cmd/, internal/ and tools/" .cache/all-cover.out

clean:
	rm -rf bin dist docs/.vitepress/dist docs/.vitepress/cache
