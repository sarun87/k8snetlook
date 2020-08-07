CGO_ENABLED=0
GOARCH=amd64
BINARY=k8snetlook

#UNAME_S := $(shell uname -s)
#ifeq (${UNAME_S},Linux)
#  GOOS=linux
#endif
#ifeq (${UNAME_S},Darwin)
#  GOOS=darwin
#endif

PWD=$(shell pwd)
BUILD_DIR=${PWD}/bin

.PHONY: all k8snetlook-linux clean k8snetlook-osx vet test

all: k8snetlook-linux

k8snetlook-linux: 
		mkdir -p ${BUILD_DIR}
		go mod tidy
		CGO_ENABLED=${CGO_ENABLED} GOARCH=${GOARCH} GOOS=linux go build -o ${BUILD_DIR}/${BINARY} ${PWD}/cmd/k8snetlook

k8snetlook-osx:
		mkdir -p ${BUILD_DIR}
		go mod tidy
		CGO_ENABLED=${CGO_ENABLED} GOARCH=${GOARCH} GOOS=darwin go build -o ${BUILD_DIR}/${BINARY}-osx ${PWD}/cmd/k8snetlook

vet:
		go mod tidy
		CGO_ENABLED=${CGO_ENABLED} GOARCH=${GOARCH} GOOS=darwin go vet ./...

test:
		go mod tidy
		CGO_ENABLED=${CGO_ENABLED} GOARCH=${GOARCH} GOOS=darwin go test -v ./...

clean: 
		rm -rf ${BUILD_DIR}
