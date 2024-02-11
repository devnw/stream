all: deps tidy fmt build test lint

op=op run --env-file="./.env" -- 

opact=op run --env-file="./.env.act" -- 

#-------------------------------------------------------------------------
# Variables
# ------------------------------------------------------------------------
env=CGO_ENABLED=0

pyenv=.venv/bin

#-------------------------------------------------------------------------
# Targets
#-------------------------------------------------------------------------
deps:
	python3 -m venv .venv

	$(pyenv)/pip install --upgrade pre-commit
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/goreleaser/goreleaser@latest

ci-test-deps:
	# install act
	if [ ! -d .act ]; then make install-act; git clone git@github.com:nektos/act.git .act; fi
	cd .act && git pull && sudo make install

test: lint 
	CGO_ENABLED=1 go test -cover -failfast -race ./...

fuzz:
	@fuzzTime=$${FUZZ_TIME:-10}; \
	files=$$(grep -r --include='**_test.go' --files-with-matches 'func Fuzz' .); \
	for file in $$files; do \
		funcs=$$(grep -o 'func Fuzz\w*' $$file | sed 's/func //'); \
		for func in $$funcs; do \
			echo "Fuzzing $$func in $$file"; \
			parentDir=$$(dirname $$file); \
			go test $$parentDir -run=$$func -fuzz=$$func -fuzztime=$${fuzzTime}s; \
			if [ $$? -ne 0 ]; then \
				echo "Fuzzing $$func in $$file failed"; \
				exit 1; \
			fi; \
		done; \
	done

bench:
	go test -bench=. -benchmem ./...

lint: tidy
	golangci-lint run
	$(pyenv)/pre-commit run --all-files	

build: update upgrade tidy lint test
	$(env) go build ./...

release: build-ci
	goreleaser release --snapshot --clean

upgrade:
	$(pyenv)/pre-commit autoupdate
	go get -u ./...

update:
	git submodule update --recursive

fmt:
	gofmt -s -w .

tidy: fmt
	go mod tidy

clean: 
	rm -rf dist
	rm -rf coverage
	rm -rf .act
	rm -rf .venv

#-------------------------------------------------------------------------
# Git targets
#-------------------------------------------------------------------------

tag:
	@latest_tag=$$(git describe --tags `git rev-list --tags --max-count=1`); \
	current_major=$$(echo $$latest_tag | cut -d. -f1); \
	current_minor=$$(echo $$latest_tag | cut -d. -f2); \
	current_patch=$$(echo $$latest_tag | cut -d. -f3); \
	next_minor=$$((current_minor + 1)); \
	default_version="$$current_major.$$next_minor.$$current_patch"; \
	read -p "Enter the version number [$$default_version]: " version; \
	version=$${version:-$$default_version}; \
	commits=$$(git log $$latest_tag..HEAD --pretty=format:"%h %s" | awk '{print "- " $$0}'); \
	git tag -a $$version -m "Release $$version" -m "$$commits"; \
	git push origin $$version

#-------------------------------------------------------------------------
# CI targets
#-------------------------------------------------------------------------
build-ci: deps
	$(env) go build ./...
	CGO_ENABLED=1 go test \
				-cover \
				-covermode=atomic \
				-coverprofile=coverage.txt \
				-failfast \
				-race ./...
	make fuzz FUZZ_TIME=10

bench-ci: build-ci
	go test -bench=. ./... | tee output.txt

release-ci: build-ci	
	goreleaser release --clean

test-ci: 
	DOCKER_HOST=$(shell docker context inspect --format='{{json .Endpoints.docker.Host}}' $(shell docker context show)) \
				$(opact) act \
					-s GIT_CREDENTIALS \
					-s GITHUB_TOKEN="$(shell gh auth token)" \
					--var GO_VERSION \
					--var ALERT_CC_USERS

#-------------------------------------------------------------------------
# Force targets
#-------------------------------------------------------------------------

FORCE: 

#-------------------------------------------------------------------------
# Phony targets
#-------------------------------------------------------------------------

.PHONY: build test lint fuzz bench fmt tidy clean release update upgrade deps translate test-act ci-test-deps
