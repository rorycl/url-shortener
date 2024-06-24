SHELL := /bin/bash
GO_VERSION := 1.22  # <1>
COVERAGE_AMT := 70  # should be 80
HEREGOPATH := $(shell go env GOPATH)
CURDIR := $(shell pwd)

build:
	go test ./... || exit 1
	go build -o url-shortener .

test:
	go test ./... -coverprofile=coverage.out
	go tool cover -func coverage.out | tee cover.rpt
	go tool cover -html=coverage.out -o cover.html

coverage-ok:
	cat cover.rpt | grep "total:" | awk '{print ((int($$3) > ${COVERAGE_AMT}) != 1) }'

clean:
	rm $$(find . -name "*cover*html" -or -name "*cover.rpt" -or -name "*coverage.out")

check: 
	test -z $$(go fmt ./...)
	test -z $$(go vet ./...)

testme:
	echo $(HEREGOPATH)

lint:
	${HEREGOPATH}/bin/golangci-lint run ./... 

module-update-tidy:
	go get -u ./...
	go mod tidy

