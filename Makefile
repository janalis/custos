# custos — developer entry points. `make verify` is the definition of done.

EA_PATH ?= $(HOME)/Sites/phpinspectionsea
RULE    ?=
GO      ?= go
BIN     := bin/custos

STUBS_REPO ?= https://github.com/JetBrains/phpstorm-stubs

FIXCHECK ?= 1

.PHONY: coverage fixcheck stubs build test vet fmt fmt-check lint bench fuzz extract rules-doc fixtures conformance cleanroom verify clean docs docs-dev

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

build:
	$(GO) build -trimpath -ldflags "-X main.version=$(VERSION)" -o $(BIN) ./cmd/custos

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

# Go sources, leaving out the docs site's installed packages.
GOFILES = $$(find . -name node_modules -prune -o -name '*.go' -print)

fmt:
	gofmt -w -s $(GOFILES)

fmt-check:
	@out=$$(gofmt -l -s $(GOFILES)); if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi

# staticcheck is pinned and run through `go run` (downloaded once into the
# module cache); set STATICCHECK= to skip it offline.
STATICCHECK ?= $(GO) run honnef.co/go/tools/cmd/staticcheck@v0.8.1

lint: vet fmt-check
	@if [ -n "$(STATICCHECK)" ]; then $(STATICCHECK) ./...; else echo "staticcheck skipped"; fi

# Benchmarks on a real file need CUSTOS_BENCH_FILE (a large PHP source).
bench:
	$(GO) test -run '^$$' -bench . -benchmem ./...
	CUSTOS_PERF=1 $(GO) test -count=1 -run TestEditLatency -v ./internal/lsp | grep -E 'p95|FAIL|SKIP|ok'

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
	$(GO) test ./internal/conformance -run 'TestOwnFixtures' -rule "$(RULE)"

# EA fixtures read from a local checkout; never copied into this repo.
conformance:
	EA_PATH="$(EA_PATH)" $(GO) test ./internal/conformance -run 'TestEA' -v -rule "$(RULE)"

# Verbatim-text scan against the local EA checkout (skipped when absent).
cleanroom:
	$(GO) run ./tools/cleanroom -ea "$(EA_PATH)"

# Apply every quick-fix offered on the real-world corpus (CUSTOS_CORPUS, a list
# of directories) and require the result to still parse (FIXCHECK=php
# additionally runs php -l on samples).
fixcheck:
	CUSTOS_FIXCHECK=$(FIXCHECK) $(GO) test ./internal/rules -run TestFixesKeepCodeParsable -v -timeout 30m

verify: lint test fixtures coverage cleanroom

# Coverage gates (the EA run and local corpora do not count):
# - every statement of internal/rules is covered by own fixtures and the rule
#   packages' tests;
# - every statement of cmd/, internal/ and tools/ is covered by the whole test
#   suite. tools/covercheck exempts only the body of a one-statement
#   `func main()` in package main (each command's os.Exit(run(...)) wrapper),
#   found from the source, so nothing has to be listed or kept in sync.
COVER_SKIP := TestEA|TestNoCrashOnCorpus|TestFixesKeepCodeParsable
coverage:
	@mkdir -p .cache
	go test ./internal/conformance ./internal/rules/... -skip '$(COVER_SKIP)' \
		-coverpkg=./internal/rules/... -coverprofile=.cache/rules-cover.out >.cache/rules-cover.log 2>&1 \
		|| { grep -E -B2 -A20 '^(--- FAIL|FAIL|panic:)' .cache/rules-cover.log; exit 1; }
	@$(GO) run ./tools/covercheck -what internal/rules .cache/rules-cover.out
	go test ./... -skip '$(COVER_SKIP)' -coverpkg=./cmd/...,./internal/...,./tools/... -coverprofile=.cache/all-cover.out >.cache/all-cover.log 2>&1 \
		|| { grep -E -B2 -A20 '^(--- FAIL|FAIL|panic:)' .cache/all-cover.log; exit 1; }
	@$(GO) run ./tools/covercheck -what "cmd/, internal/ and tools/" .cache/all-cover.out

clean:
	rm -rf bin dist docs/.vitepress/dist docs/.vitepress/cache
