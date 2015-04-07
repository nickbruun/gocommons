PACKAGES := \
	logging \
	zkutils \
	distributed/leadership
DEPENDENCIES := \
	github.com/Sirupsen/logrus \
	github.com/samuel/go-zookeeper/zk

GOPATH ?= $(shell pwd)
GOPATH_FIRST := $(firstword $(subst :, ,$(GOPATH)))
export GOPATH

DEPENDENCIES_DIRS := $(addprefix $(GOPATH_FIRST)/src/, $(DEPENDENCIES))
GO_SRC := $(wildcard *.go) $(wildcard $(PACKAGES:%=%/*.go)) $(wildcard $(PACKAGES:%=%/**/*.go))

$(GOPATH_FIRST)/src/%:
	go get $(@:$(GOPATH_FIRST)/src/%=%)

test: $(DEPENDENCIES_DIRS)
	go test -v $(PACKAGES:%=./%)

format:
	gofmt -l -w $(GO_SRC)

.PHONY: test format
