# VERSION is autogenerated in a format of
# {{ latest tag }}
# +{{ current commit tag }}  if there have been changes since latest tag
# +dirty                     if there are uncommited changes
GITTAGS = $(shell git describe --tags 2>/dev/null | cut -d '-' -f 1,3 | tr '-' '+' | tr -d 'g')
DIRTY = $(shell git diff-index --quiet HEAD -- || printf "+dirty")
ifeq ($(GITTAGS),)
VERSION = +$(shell git log -1 --pretty=format:"%h")
else
VERSION = $(GITTAGS)
endif

.PHONY: build
build:
	mkdir -p bin/ && go build \
		-ldflags "-X \"github.com/m-horky/insights-client-next/internal/constants.Version=$(VERSION)$(DIRTY)\"" \
		-o bin/ ./...
