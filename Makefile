CGO_ENABLED=0
GOARCH=amd64
BINARY=k8snetlook
TAG=v0.3

UNAME_S := $(shell uname -s)
ifeq (${UNAME_S},Linux)
  GOOS=linux
endif
ifeq (${UNAME_S},Darwin)
  GOOS=darwin
endif

PWD=$(shell pwd)
BUILD_DIR=${PWD}/bin

.PHONY: all k8snetlook-linux clean k8snetlook-osx vet test release docker-image

all: k8snetlook-linux

k8snetlook-linux: 
		mkdir -p ${BUILD_DIR}
		go mod tidy
		CGO_ENABLED=${CGO_ENABLED} GOARCH=${GOARCH} GOOS=linux go build -ldflags="-s -w" -o ${BUILD_DIR}/${BINARY} ${PWD}/cmd/k8snetlook

k8snetlook-osx:
		mkdir -p ${BUILD_DIR}
		go mod tidy
		CGO_ENABLED=${CGO_ENABLED} GOARCH=${GOARCH} GOOS=darwin go build -ldflags="-s -w" -o ${BUILD_DIR}/${BINARY}-osx ${PWD}/cmd/k8snetlook

vet:
		go mod tidy
		CGO_ENABLED=${CGO_ENABLED} GOARCH=${GOARCH} GOOS=${GOOS} go vet ./...

test:
		go mod tidy
		sudo CGO_ENABLED=${CGO_ENABLED} GOARCH=${GOARCH} GOOS=${GOOS} go test -v ./...

clean: 
		rm -rf ${BUILD_DIR}

release: k8snetlook-linux
		cd ${BUILD_DIR} && \
		upx ${BINARY}

docker-image: release
		docker build --network host -t sarun87/k8snetlook:${TAG} .
