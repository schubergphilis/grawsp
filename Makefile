SHELL := /bin/bash

ROOT_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
SOURCE_DIR := $(abspath ${ROOT_DIR}/src)
TESTS_DIR := $(abspath ${ROOT_DIR}/tests)

PROJECT_BUILD_DATE ?= $(shell date --rfc-3339=seconds)
PROJECT_COMMIT ?= $(shell git rev-parse HEAD)
PROJECT_NAME ?= $(error PROJECT_NAME is not set)
PROJECT_VERSION ?= $(strip \
	$(if \
		$(shell git rev-list --tags --max-count=1), \
		$(shell git describe --tags `git rev-list --tags --max-count=1`), \
		$(shell git rev-parse --short HEAD) \
	) \
)

.PHONY: all
all: lint scan build test

.PHONY: build
build:
	@poetry build

.PHONY: clean
clean:
	rm -f "$(ROOT_DIR)/.ruff_cache"

.PHONY: env
env:
	@direnv allow "$(ROOT_DIR)"

.PHONY: lint
lint:
	@poetry run ruff check --fix

.PHONY: release
release:
	@echo "Not yet implemented."

.PHONY: scan
scan:
	@poetry run bandit -r "$(SOURCE_DIR)"

.PHONY: reset
reset: clean
	@rm -rf "$(ROOT_DIR)/.venv"

.PHONY: test
test:
	@poetry run pytest -ra -q
