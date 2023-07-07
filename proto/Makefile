SHELL=/bin/bash -o pipefail

export GO111MODULE        := on
export PWD                := $(shell pwd)
export GOBIN              := $(PWD)/.bin/
export PATH               := ${GOBIN}:${PATH}

PROTOC_GEN_GO := $(GOBIN)/protoc-gen-go

$(PROTOC_GEN_GO):
	go install google.golang.org/protobuf/cmd/protoc-gen-go

divera-alarm.pb.go: divera-alarm.proto | $(PROTOC_GEN_GO)
	protoc --go_out=. --go_opt=paths=source_relative divera-alarm.proto

.PHONY: compile
compile: divera-alarm.pb.go
