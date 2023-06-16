SHELL:=/bin/bash

changelog_args=-o CHANGELOG.md -tag-filter-pattern '^v'

.PHONY: changelog
changelog:
ifdef version
	$(eval changelog_args=--next-tag $(version) $(changelog_args))
endif
	git-chglog $(changelog_args)

.PHONY: lint
lint:
	golangci-lint run --print-issued-lines=false --exclude-use-default=false --fix --timeout=3m

.PHONY: test-only
test-only: ; $(info $(M) start unit testing...) @
	@go test $$(go list ./... | grep -v /mocks/) --race -v -short -coverprofile=profile.cov
	@echo "\n*****************************"
	@echo "**  TOTAL COVERAGE: $$(go tool cover -func profile.cov | grep total | grep -Eo '[0-9]+\.[0-9]+')%  **"
	@echo "*****************************\n"

.PHONY: test
test: lint test-only
