VERSION?=0.0.0

.PHONY: build
build:
	mkdir -p bin/ && \
	go build \
		-ldflags "-X \"github.com/m-horky/insights-client-next/internal.Version=$(VERSION)\"" \
		-o bin/ ./...

.PHONY: check
check:
	goimports -w -l .
	gofmt -w -l .
