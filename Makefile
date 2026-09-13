# Makefile for terraform-provider-atlas.
#
# `test` runs the fast, hermetic unit tests (the client request/error mapping);
# `testacc` runs the framework acceptance tests, which stand up the provider
# in-process and drive real CRUD against a live Atlas instance.

BINARY  := terraform-provider-atlas
TIMEOUT := 120m

.PHONY: build vet fmt fmtcheck test testacc

## build: compile the provider binary.
build:
	go build -o $(BINARY) .

## vet: run go vet across the module.
vet:
	go vet ./...

## fmt: rewrite sources with gofmt.
fmt:
	gofmt -w .

## fmtcheck: fail if any file is not gofmt-clean.
fmtcheck:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "These files are not gofmt-clean:"; echo "$$unformatted"; exit 1; \
	fi

## test: run unit tests only. Acceptance tests self-skip without TF_ACC.
test:
	go test ./... -v

## testacc: run the acceptance tests against a live instance.
##
## These CREATE AND DESTROY REAL OBJECTS in the target Atlas instance, so point
## them at a disposable test instance, never production. Required environment:
##
##   TF_ACC=1              set by this target; without it every acc test is skipped
##   ATLAS_SECRET_KEY      a real instance secret key (sk_...) for the test instance
##   ATLAS_API_URL         the Backend API origin of that instance
##   ATLAS_TEST_USER_ID    an existing user id (only the atlas_organization tests
##                         need it; those tests skip when it is unset)
testacc:
	TF_ACC=1 go test ./... -v -timeout $(TIMEOUT)
