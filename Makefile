SHELL := /bin/bash

ROOT_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
DIST_DIR := $(abspath ${ROOT_DIR}/dist)
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

PYPI_UPLOAD_USERNAME ?=
PYPI_UPLOAD_PASSWORD ?=

.PHONY: all
all: lint scan build test

.PHONY: build
build:
	@poetry build

.PHONY: clean
clean:
	@rm -rf "$(ROOT_DIR)/.ruff_cache"
	@find "$(ROOT_DIR)" -type d -name "__pycache__" -exec rm -rf {} +
	@rm -f "$(DIST_DIR)"/*

.PHONY: env
env:
	@direnv allow "$(ROOT_DIR)"

.PHONY: lint
lint:
	@poetry run ruff check --fix

.PHONY: release
release:
	@poetry publish \
		-u "$(PYPI_UPLOAD_USERNAME)" \
		-p "$(PYPI_UPLOAD_PASSWORD)"

.PHONY: scan
scan:
	@poetry run bandit -r "$(SOURCE_DIR)"

.PHONY: reset
reset: clean
	@rm -rf "$(ROOT_DIR)/.venv"

.PHONY: test
test:
	@poetry run pytest -ra -q
