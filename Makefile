all: build tidy lint fmt test

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



lint: tidy
	golangci-lint run
	$(pyenv)/pre-commit run --all-files	

build: update upgrade tidy lint test
	$(env) go build ./...

release-dev: build-ci
	goreleaser release --rm-dist --snapshot

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

#-------------------------------------------------------------------------
# CI targets
#-------------------------------------------------------------------------
test-ci: deps tidy lint
	CGO_ENABLED=1 go test \
				-cover \
				-covermode=atomic \
				-coverprofile=coverage.txt \
				-failfast \
				-race ./...
	make fuzz FUZZ_TIME=10

build-ci: test-ci
	$(env) go build ./...


release-ci: build-ci
	goreleaser release --rm-dist	

#-------------------------------------------------------------------------
# Force targets
#-------------------------------------------------------------------------

FORCE: 

#-------------------------------------------------------------------------
# Phony targets
#-------------------------------------------------------------------------

.PHONY: build test lint fuzz
