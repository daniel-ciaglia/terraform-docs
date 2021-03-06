NAME        := terraform-docs
VENDOR      := daniel-ciaglia
DESCRIPTION := Generate docs from Terraform modules / a forked version
MAINTAINER  := Daniel Ciaglia <daniel.ciaglia@kreuzwerker.de>
URL         := https://github.com/$(VENDOR)/$(NAME)
LICENSE     := MIT

VERSION     := $(shell cat ./VERSION)

GOBUILD     := go build -ldflags "-X main.version=$(VERSION)"
GOPKGS      := $(shell go list ./... | grep -v /vendor)

TOKEN = $(shell cat ~/.github.token)

.PHONY: all
all: clean deps test build

.PHONY: authors
authors:
	git log --all --format='%aN <%aE>' | sort -u | egrep -v noreply > AUTHORS

.PHONY: build
build: authors build-darwin-amd64
	$(eval SHA256 = $(firstword $(shell shasum -p -a 256 bin/darwin-amd64/$(NAME))))
	@echo "$(SHA256) bin/darwin-amd64/$(NAME)"

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o bin/darwin-amd64/$(NAME)

.PHONY: clean
clean:
	rm -rf ./bin

.PHONY: changelog
changelog:
	git-chglog -o CHANGELOG.md

.PHONY: deps
deps:
	GO111MODULE=on go mod vendor

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: release
release:
	git tag -a v$(VERSION) -m v$(VERSION) -f && git push --tags -f
	@github-release release --user $(VENDOR) --repo $(NAME) --tag v$(VERSION) -s $(TOKEN)
	@github-release upload --user $(VENDOR) --repo $(NAME) --tag v$(VERSION) -s $(TOKEN) --name $(NAME)-darwin-amd64 --file bin/darwin-amd64/$(NAME)

.PHONY: retract
retract:
	@github-release delete --user $(VENDOR) --repo $(NAME) --tag v$(VERSION) -s $(TOKEN)

.PHONY: test
test:
	go test -v $(GOPKGS)
