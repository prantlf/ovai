RELFLAGS=
ifeq (1,${RELEASE})
	RELFLAGS:=-trimpath -gcflags=all="-B" -ldflags="-s -w -buildid="
endif

all: lint build test

lint:
	go vet ./...

format:
	go fmt ./...

build:
	go build $(RELFLAGS) -o ovai ./cmd/ovai/main.go

test:
	go test ./internal/...

bench::
	cd bench && go test -benchmem -bench=.

clean:
	go clean
	rm -f ovai

docker: docker-lint docker-build

docker-clean:
	docker image rm ovai

docker-lint:
	docker run --rm -i \
		-v ${PWD}/.hadolint.yaml:/bin/hadolint.yaml \
		-e XDG_CONFIG_HOME=/bin hadolint/hadolint \
		< Dockerfile

docker-build:
	docker build -t ovai .

docker-enter:
	docker run --rm -it -p 22434:22434 --entrypoint sh \
		-v ${PWD}/google-account.json:/google-account.json \
		-v ${PWD}/model-defaults.json:/model-defaults.json \
		ovai

docker-start:
	docker run --rm -dt -p 22434:22434 --name ovai \
		-v ${PWD}/google-account.json:/google-account.json \
		-v ${PWD}/model-defaults.json:/model-defaults.json \
		ovai

docker-kill:
	docker container kill ovai

docker-log:
	docker logs ovai

docker-up:
	IMAGE_HUB= docker compose -f docker-compose.yml up -d --wait

docker-down:
	IMAGE_HUB= docker compose -f docker-compose.yml down
