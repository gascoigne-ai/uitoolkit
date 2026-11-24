.PHONY: check bench fuzz fmt-check vet test

check: fmt-check vet test

fmt-check:
	@test -z "$$(gofmt -l .)" || { gofmt -l .; exit 1; }

vet:
	go vet ./...

test:
	go test -race ./...

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
	./desktop/keyrepeat:FuzzParseWindows
fuzz:
	@for target in $(FUZZ_TARGETS); do \
		package=$${target%%:*}; name=$${target##*:}; \
		echo "fuzz $$name ($$package)"; \
		go test $$package -run '^$$' -fuzz "^$$name$$" -fuzztime $(FUZZTIME) || exit 1; \
	done
