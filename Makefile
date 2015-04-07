PACKAGES := \
	logging \
	zkutils
DEPENDENCIES := \
	github.com/Sirupsen/logrus \
	github.com/samuel/go-zookeeper/zk

GOPATH ?= $(shell pwd)
export GOPATH

DEPENDENCIES_DIRS := $(addprefix $(GOPATH)/src/, $(DEPENDENCIES))

#$(GOPATH)/src/%:
#	go get $(@:$(GOPATH)/src/%=%)

test:
	@echo $(DEPENDENCIES_DIRS)
	go test -v $(PACKAGES:%=./%)

format:
	gofmt -l -w $(wildcard $(PACKAGES:%=%/*.go)) $(wildcard $(PACKAGES:%=%/**/*.go))

.PHONY: test format
