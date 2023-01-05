.PHONY: all build release

IMAGE=ghcr.io/dnachev/wg-nc

all: build

build-alpine:
	CGO_ENABLED=0 GOOS=linux go test
	CGO_ENABLED=0 GOOS=linux go build -a -o ./bin/wg-nc ./main.go

build:
	@docker build --tag=${IMAGE} .

debug:
	@docker run -it --entrypoint=sh ${IMAGE}

release: build
	@echo "Tag image with version $(version)"
	@docker tag ${IMAGE} ${IMAGE}:$(version)

push: release
	@docker push ${IMAGE}
	@docker push ${IMAGE}:$(version)
