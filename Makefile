MAIN_PACKAGE_PATH := ./
BINARY_NAME := beampaw

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

.PHONY: no-dirty
no-dirty:
	git diff --exit-code > /dev/null

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## tidy: format code and tidy modfile
.PHONY: tidy
tidy:
	go fmt ./...
	go mod tidy -v

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## init/vars: configure the project variables
.PHONY: init/vars
init:
	cp .env.example .env
	ssh-keygen -t rsa -b 4096 -f id_rsa -q -N ''

## init/modules: install dependencies
.PHONY: init/modules
init/modules:
	npm i
	go mod tidy

## build: build the application
.PHONY: build
build:
	npx tailwindcss -i ./styles.css -o ./public/index.css --minify
	go build -o=/tmp/bin/${BINARY_NAME} ${MAIN_PACKAGE_PATH}

## build/docker: build the application for docker
.PHONY: build/docker
build/docker:
	docker build -t ${BINARY_NAME}:latest .

## run: run the  application
.PHONY: run
run: build
	/tmp/bin/${BINARY_NAME}

## run/live: run the application with reloading on file changes with tailwind watch
.PHONY: run/watch
run/watch:
	go run github.com/cosmtrek/air@v1.43.0 \
        --build.cmd "make build" --build.bin "/tmp/bin/${BINARY_NAME}" --build.delay "100" \
        --build.exclude_dir "node_modules" \
        --build.include_ext "go, tpl, tmpl" \
        --misc.clean_on_exit "true"


## tailwind-watch: tailwindcss watch mode
.PHONY: tailwind-watch
tailwind-watch:
	npx tailwindcss -i ./styles.css -o ./public/index.css --watch

# ==================================================================================== #
# OPERATIONS
# ==================================================================================== #

## production/deploy: deploy the application to production
.PHONY: production/deploy
production/deploy: confirm tidy no-dirty
	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=/tmp/bin/linux_amd64/${BINARY_NAME} ${MAIN_PACKAGE_PATH}
	echo "output to /tmp/bin/linux_amd64/${BINARY_NAME}"
	echo "this script is not yet complete"
