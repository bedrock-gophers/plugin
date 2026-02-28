SHELL := /usr/bin/env bash

.PHONY: help generate fmt test plugins plugins-docker dev-plugins run run-remote run-server dev

help:
	@echo "Targets:"
	@echo "  make generate     - regenerate internal code"
	@echo "  make fmt          - format Go source files"
	@echo "  make test         - run go test ./..."
	@echo "  make plugins      - build example plugins (Go .so + C# publish output) into ./plugins"
	@echo "  make plugins-docker - build example plugins using Docker only"
	@echo "  make dev-plugins  - build plugins and print in-game reload reminder"
	@echo "  make run          - run ./start.sh using local host source (USE_LOCAL_HOST_WORKTREE=1)"
	@echo "  make run-remote   - run ./start.sh using cached/fetched host source"
	@echo "  make run-server   - run server directly (go run ./cmd)"
	@echo "  make dev          - generate + fmt + test + plugins"

generate:
	cd plugin/internal/ctxkey && go generate ./...
	cd plugin/internal/hostapi && go generate ./...
	cd plugin/internal/catalog && go generate ./...

fmt:
	gofmt -w $$(find . -type f -name '*.go')

test:
	go test ./...

plugins:
	./build_plugins.sh

plugins-docker:
	./build_plugins_docker.sh

dev-plugins: plugins
	@echo "plugins rebuilt. run /pl reload all in-game."

run:
	USE_LOCAL_HOST_WORKTREE=1 ./start.sh

run-remote:
	./start.sh

run-server:
	go run ./cmd

dev: generate fmt test plugins
	@echo "dev cycle complete. if server is running, run /pl reload all in-game."
