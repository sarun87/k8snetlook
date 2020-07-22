CGO_ENABLED=0
GOARCH=amd64
BINARY=k8snetlook

UNAME_S := $(shell uname -s)
ifeq (${UNAME_S},Linux)
  GOOS=linux
endif
ifeq (${UNAME_S},Darwin)
  GOOS=darwin
endif

PWD=$(shell pwd)
BUILD_DIR=${PWD}/bin

.PHONY: all k8snetlook clean

all: k8snetlook

k8snetlook: 
		mkdir -p ${BUILD_DIR}
		go mod tidy
		CGO_ENABLED=${CGO_ENABLED} GOARCH=${GOARCH} GOOS=${GOOS} go build -o ${BUILD_DIR}/${BINARY} ${PWD}/cmd/k8snetlook

clean: 
		rm -rf ${BUILD_DIR}