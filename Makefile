GOPATH:=$(shell go env GOPATH)

ifdef DOMAIN
    DOMAIN:=-${DOMAIN}
endif

.PHONY: run
run:
	MICRO_LOG_LEVEL=debug \
	MICRO_REGISTRY=consul \
	MICRO_BROKER=kafka \
	MICRO_BROKER_ADDRESS=127.0.0.1:9091 \
	MICRO_SERVER_VERSION=latest \
	MICRO_SERVER_NAME=brazn.web.socket${DOMAIN} \
	go run main.go plugin.go

.PHONY: proto
proto:

.PHONY: build
build: proto
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${PWD}/bin/app ./main.go ./plugin.go

	chmod +x ${PWD}/bin/app

.PHONY: docker
docker:
	docker build . -t brazn/websocket-srv:latest
