# custos — developer entry points. `make verify` is the definition of done.

EA_PATH ?= $(HOME)/Sites/phpinspectionsea
RULE    ?=
GO      ?= go
BIN     := bin/custos

STUBS_REPO ?= https://github.com/JetBrains/phpstorm-stubs

FIXCHECK ?= 1

.PHONY: fixcheck stubs build test vet fmt fmt-check lint bench fuzz extract rules-doc fixtures conformance cleanroom verify clean

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

build:
	$(GO) build -trimpath -ldflags "-X main.version=$(VERSION)" -o $(BIN) ./cmd/custos

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

fmt:
	gofmt -w -s .

fmt-check:
	@out=$$(gofmt -l -s .); if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi

# staticcheck is pinned and run through `go run` (downloaded once into the
# module cache); set STATICCHECK= to skip it offline.
STATICCHECK ?= $(GO) run honnef.co/go/tools/cmd/staticcheck@v0.8.1

lint: vet fmt-check
	@if [ -n "$(STATICCHECK)" ]; then $(STATICCHECK) ./...; else echo "staticcheck skipped"; fi

# Benchmarks on a real file need CUSTOS_BENCH_FILE (a large PHP source).
bench:
	$(GO) test -run '^$$' -bench . -benchmem ./...

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

verify: lint test fixtures cleanroom

clean:
	rm -rf bin dist
