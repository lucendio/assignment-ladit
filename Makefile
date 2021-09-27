SHELL := /usr/bin/env bash -euo pipefail

.DEFAULT_GOAL := default

MKFILE_DIR = $(abspath $(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
CWD ?= $(CURDIR)
LOCAL_DIR := $(abspath $(MKFILE_DIR)/.local)

BIN_DIR	  := $(LOCAL_DIR)/bin


export PATH := $(BIN_DIR):$(PATH)

export GOPATH = $(LOCAL_DIR)/cache/go
export GO111MODULE = on



default: clean install build run


.PHONY: install
install: $(GOPATH)
$(GOPATH):
	mkdir -p $(GOPATH)
	cd $(MKFILE_DIR)/src \
		&& go get -t

.PHONY: run
run:
	# encoded: c2Vuc2l0aXZlCg==
	cd $(MKFILE_DIR)/src \
	&& ACCESS_TOKEN='sensitive' go run .

.PHONY: build $(BIN_DIR)/artifact.bin
build: $(BIN_DIR)/artifact.bin
$(BIN_DIR)/artifact.bin:
	cd $(MKFILE_DIR)/src \
	&& go build \
		-o $(@) \
		$(MKFILE_DIR)/src/*.go

.PHONY: image
image:
	docker build \
		--file $(MKFILE_DIR)/Containerfile \
		--tag 'blocksvc' \
		$(MKFILE_DIR)


.PHONY: test
.SILENT: test
test:
	cd $(MKFILE_DIR)/src \
	&& go test \
		-race \
		-v \
		$(MKFILE_DIR)/src/...


.PHONY: exec
exec:
	chmod +x $(BIN_DIR)/artifact.bin
	exec $(BIN_DIR)/artifact.bin


.PHONY: clean
clean:
	rm -rf \
		$(LOCAL_DIR)


.PHONY: tf-%
tf-%: export TF_DATA_DIR = $(LOCAL_DIR)/terraform
tf-%:
	terraform \
		-chdir='$(MKFILE_DIR)/terraform' \
		$(*)
