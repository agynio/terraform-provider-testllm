HOSTNAME=registry.terraform.io
NAMESPACE=agynio
NAME=testllm
BINARY=terraform-provider-${NAME}
VERSION?=0.0.1
OS?=$(shell go env GOOS)
ARCH?=$(shell go env GOARCH)

.PHONY: build install test testacc lint generate fmt clean

build:
	go build -o ${BINARY} ./...

install:
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS}_${ARCH}
	go build -o ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS}_${ARCH}/${BINARY} ./...

test:
	go test ./...

testacc:
	TF_ACC=1 go test ./... -v -timeout 120m

lint:
	go vet ./...

generate:
	cd tools && go generate ./...

fmt:
	gofmt -w .

clean:
	rm -f ${BINARY}
