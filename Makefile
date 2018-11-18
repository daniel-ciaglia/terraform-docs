NAME        := terraform-docs
VENDOR      := daniel-ciaglia
DESCRIPTION := Generate docs from Terraform modules / a forked version
MAINTAINER  := Daniel Ciaglia <daniel.ciaglia@kreuzwerker.de>
URL         := https://github.com/$(VENDOR)/$(NAME)
LICENSE     := MIT

VERSION     := $(shell cat ./VERSION)

GOBUILD     := go build -ldflags "-X main.version=$(VERSION)"
GOPKGS      := $(shell go list ./... | grep -v /vendor)

TOKEN = $(shell cat .token)

.PHONY: all
all: clean deps test build

.PHONY: authors
authors:
	git log --all --format='%aN <%aE>' | sort -u | egrep -v noreply > AUTHORS

.PHONY: build
build: authors build-darwin-amd64
	$(eval SHA256 = $(firstword $(shell shasum -p -a 256 bin/darwin-amd64/$(NAME))))
	sed  -i '' 's/.*sha256.*/  sha256 "$(SHA256)"/' terraform-docs.rb

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o bin/darwin-amd64/$(NAME)

.PHONY: clean
clean:
	rm -rf ./bin

.PHONY: lint
lint:
	gometalinter --config gometalinter.json ./...

.PHONY: deps
deps:
	dep ensure

.PHONY: release
release:
	git tag -a v$(VERSION) -m v$(VERSION) && git push --tags
	github-release release --user $(VENDOR) --repo $(NAME) --tag $(VERSION) -s $(TOKEN)
	github-release upload --user $(VENDOR) --repo $(NAME) --tag $(VERSION) -s $(TOKEN) --name $(NAME)-darwin-amd64 --file bin/darwin-amd64/$(NAME)

.PHONY: retract
retract:
	github-release delete --user $(VENDOR) --repo $(NAME) --tag $(VERSION) -s $(TOKEN)

.PHONY: test
test:
	go test -v $(GOPKGS)
