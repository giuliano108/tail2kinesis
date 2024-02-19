NS ?= giuliano108
VERSION ?= latest

IMAGE_NAME ?= tail2kinesis
CONTAINER_NAME ?= tail2kinesis
CONTAINER_INSTANCE ?= default
PORTS = 
VOLUMES = -v $(CURDIR):/src:ro

NPM_REGISTRY=https://registry.npmjs.org/

.PHONY: build shell run rm

docker-build: Dockerfile
	docker build -t $(NS)/$(IMAGE_NAME):$(VERSION) --build-arg NPM_REGISTRY=$(NPM_REGISTRY) -f Dockerfile .

docker-shell:
	docker run --rm --name $(CONTAINER_NAME)-$(CONTAINER_INSTANCE) -i -t $(PORTS) $(VOLUMES) $(ENV) $(NS)/$(IMAGE_NAME):$(VERSION) /bin/bash

docker-run:
	docker run --rm --name $(CONTAINER_NAME)-$(CONTAINER_INSTANCE) $(PORTS) $(VOLUMES) $(ENV) $(NS)/$(IMAGE_NAME):$(VERSION)

docker-rm:
	docker rm $(CONTAINER_NAME)-$(CONTAINER_INSTANCE)

default: docker-build

build:
	go build .

test:
	go test -test.v ./...
