VERSION = 4.0.0a0
# When building from commit that is not tagged, include Git hash.
GITHASH = $(shell git log -1 --pretty=format:"%h")
GITTAGS = $(shell git describe --tags 2>/dev/null)
ifeq ($(GITTAGS),)
RICH_VERSION = $(VERSION)+$(GITHASH)
else
RICH_VERSION = $(VERSION)
endif

.PHONY: build
build:
	mkdir -p bin/ && \
	go build \
		-ldflags "-X \"github.com/m-horky/insights-client-next/internal/constants.Version=$(RICH_VERSION)\"" \
		-o bin/ ./...
