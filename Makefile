SHELL := '/bin/bash'

.DEFAULT_GOAL := default

MKFILE_DIR = $(abspath $(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
CWD ?= $(CURDIR)
LOCAL_DIR := $(abspath $(MKFILE_DIR)/.local)

BIN_DIR	  := $(LOCAL_DIR)/bin


export GOPATH = $(LOCAL_DIR)/cache/go
export GO111MODULE = on



default: clean install build run


.PHONY: install
install: $(GOPATH)
$(GOPATH):
	mkdir -p $(GOPATH)
	cd $(MKFILE_DIR)/src \
		&& go get -d

.PHONY: run
run:
	cd $(MKFILE_DIR)/src \
	&& go run \
		$(MKFILE_DIR)/src/*.go

.PHONY: build $(BIN_DIR)/artifact.bin
build: $(BIN_DIR)/artifact.bin
$(BIN_DIR)/artifact.bin:
	cd $(MKFILE_DIR)/src \
	&& go build \
		-o $(@) \
		$(MKFILE_DIR)/src/*.go


.PHONY: test
.SILENT: test
test:
	go clean -testcache
	go test \
		-race \
		-v $(MKFILE_DIR)/src/...


.PHONY: exec
exec:
	chmod +x $(BIN_DIR)/artifact.bin
	exec $(BIN_DIR)/artifact.bin


.PHONY: clean
clean:
	rm -rf \
		$(LOCAL_DIR)
