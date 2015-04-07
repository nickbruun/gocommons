PACKAGES := \
	logging \
	zkutils
DEPENDENCIES := \
	github.com/Sirupsen/logrus \
	github.com/samuel/go-zookeeper/zk

GOPATH ?= $(shell pwd)
GOPATH_FIRST := $(firstword $(subst :, ,$(GOPATH)))
export GOPATH

DEPENDENCIES_DIRS := $(addprefix $(GOPATH_FIRST)/src/, $(DEPENDENCIES))

$(GOPATH_FIRST)/src/%:
	go get $(@:$(GOPATH_FIRST)/src/%=%)

test: $(DEPENDENCIES_DIRS)
	go test -v $(PACKAGES:%=./%)

format:
	gofmt -l -w $(wildcard $(PACKAGES:%=%/*.go)) $(wildcard $(PACKAGES:%=%/**/*.go))

.PHONY: test format
