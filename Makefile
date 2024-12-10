DOCKER_ACCOUNT = grizzlybite
PROGRAM_NAME = katok

PKG_PATH = github.com/grizzlybite/katok
COMMIT=$(shell git rev-parse --short HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
TAG=$(shell git describe --tags |cut -d- -f1)

LDFLAGS = -ldflags "-X ${PKG_PATH}/internal/version.gitTag=${TAG} -X ${PKG_PATH}/internal/version.gitCommit=${COMMIT} -X ${PKG_PATH}/internal/version.gitBranch=${BRANCH}"

.PHONY: help clean dep build install uninstall

.DEFAULT_GOAL := help

help: ## Display this help screen.
	@echo "Makefile available targets:"
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  * \033[36m%-15s\033[0m %s\n", $$1, $$2}'

clean: ## Clean build directory.
	rm -f ./bin/${PROGRAM_NAME}
	rmdir ./bin

dep: ## Download the dependencies.
	go mod download

lint: dep ## Lint the source files
	golangci-lint run --timeout 5m -E golint -e '(struct field|type|method|func) [a-zA-Z`]+ should be [a-zA-Z`]+'
	## gosec -quiet ./...

build: dep ## Build katok executable.
	mkdir -p ./bin
	CGO_ENABLED=0 go build ${LDFLAGS} -o bin/${PROGRAM_NAME} ./main.go

install: ## Install katok executable into /usr/bin directory.
	install -pm 755 bin/${PROGRAM_NAME} /usr/bin/${PROGRAM_NAME}

uninstall: ## Uninstall katok executable from /usr/bin directory.
	rm -f /usr/bin/${PROGRAM_NAME}

docker-build: ## Build docker image
	docker build -t ${DOCKER_ACCOUNT}/${PROGRAM_NAME}:${TAG} .
	docker tag ${DOCKER_ACCOUNT}/${PROGRAM_NAME}:${TAG} ${DOCKER_ACCOUNT}/${PROGRAM_NAME}:latest
	docker image prune --force --filter label=stage=intermediate

docker-push: ## Push docker image to registry
	docker push ${DOCKER_ACCOUNT}/${PROGRAM_NAME}:${TAG}
	docker push ${DOCKER_ACCOUNT}/${PROGRAM_NAME}:latest
