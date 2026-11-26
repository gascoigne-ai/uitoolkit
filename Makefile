.PHONY: check bench fuzz fmt-check vet test deps-check

check: fmt-check vet test deps-check

fmt-check:
	@test -z "$$(gofmt -l .)" || { gofmt -l .; exit 1; }

# The root module is checked with GOWORK=off to prove it stands alone; go.work
# is for local development only. `./...` in workspace mode would not reach the
# nested modules, so they are named: the renderer and the examples.
vet:
	GOWORK=off go vet ./...
	go vet ./renderer/ebiten/... ./examples/...

test:
	GOWORK=off go test -race ./...
	go test -race ./renderer/ebiten/... ./examples/...
# Allocation counts hold only without the race detector (alloc_test.go).
	go test -run '^TestTextCallsAllocate' ./renderer/ebiten

# The root module must not depend on a GPU library, tests included, must be
# tidy and must build without cgo.
deps-check:
	@if GOWORK=off go list -deps -test ./... | grep -E 'hajimehoshi|wgpu'; then \
		echo "root module depends on a GPU library" >&2; exit 1; \
	fi
	GOWORK=off go mod tidy -diff
	CGO_ENABLED=0 GOWORK=off go build ./...

# Frame-cost benchmarks (CPU side); not part of check. Compare two runs with
# benchstat when installed.
BENCHCOUNT ?= 6
bench:
	go test -run '^$$' -bench . -benchmem -count $(BENCHCOUNT) ./...

# Short run of every fuzz target (package:Target pairs below); not part of
# check. A crasher lands in the package's testdata/fuzz and becomes a
# regression test: fix it, keep the file.
FUZZTIME ?= 30s
FUZZ_TARGETS = \
	./markup:FuzzParse \
	./widget:FuzzEdit \
	./widget:FuzzFind \
	./widget:FuzzParseSVG \
	./widget:FuzzWrapRunes \
	./theming:FuzzParseTheme \
	./desktop/displayscale:FuzzParseMonitors \
	./desktop/keyrepeat:FuzzParseHyprland \
	./desktop/keyrepeat:FuzzParseGSettings \
	./desktop/keyrepeat:FuzzParseXset \
	./desktop/keyrepeat:FuzzParseMac \
	./desktop/keyrepeat:FuzzParseWindows \
	./render/text:FuzzNewFontSource
fuzz:
	@for target in $(FUZZ_TARGETS); do \
		package=$${target%%:*}; name=$${target##*:}; \
		echo "fuzz $$name ($$package)"; \
		go test $$package -run '^$$' -fuzz "^$$name$$" -fuzztime $(FUZZTIME) || exit 1; \
	done
