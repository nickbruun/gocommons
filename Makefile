PACKAGES := \
	logging \
	zkutils
DEPENDENCIES := \
	github.com/Sirupsen/logrus \
	github.com/samuel/go-zookeeper/zk

DEPENDENCIES_DIRS := $(addprefix src/, $(DEPENDENCIES))

export GOPATH=$(shell pwd)

$(DEPENDENCIES_DIRS):
	go get $(@:src/%=%)

test: $(DEPENDENCIES_DIRS)
	go test -v $(PACKAGES:%=./%)

format:
	gofmt -l -w $(wildcard $(PACKAGES:%=%/*.go)) $(wildcard $(PACKAGES:%=%/**/*.go))

.PHONY: test format
