# VERSION is generated in a format of {{ latest tag }}+g{{ current commit tag }}
GITTAGS = $(shell git describe --tags 2>/dev/null | cut -d '-' -f 1,3 | tr '-' '+')
ifeq ($(GITTAGS),)
VERSION = +g$(shell git log -1 --pretty=format:"%h")
else
VERSION = $(GITTAGS)
endif

.PHONY: build
build:
	mkdir -p bin/ && go build \
		-ldflags "-X \"github.com/m-horky/insights-client-next/internal/constants.Version=$(VERSION)\"" \
		-o bin/ ./...
