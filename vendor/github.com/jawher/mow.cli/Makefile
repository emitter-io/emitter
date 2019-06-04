test:
	go test ./...

test-cover:
	go list -f '{{if len .TestGoFiles}}"go test -coverprofile={{.Dir}}/.coverprofile {{.ImportPath}}"{{end}}' ./... | xargs -L 1 sh -c
	gover
	goveralls -coverprofile=gover.coverprofile -service=travis-ci

check: readmecheck
	bin/golangci-lint run

doc:
	autoreadme -f

readmecheck:
	sed '$ d' README.md > README.original.md
	autoreadme -f
	sed '$ d' README.md > README.generated.md
	diff README.generated.md README.original.md

setup:
	go get golang.org/x/tools/cmd/cover
	go get github.com/mattn/goveralls
	go get github.com/modocache/gover
	go get github.com/divan/autoreadme
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s v1.16.0

.PHONY: test check lint vet fmtcheck ineffassign readmecheck
