VERSION=0.0.7
LDFLAGS=-ldflags "-w -s -X main.version=${VERSION}"
GO111MODULE=on

all: mackerel-plugin-mysql-lite

.PHONY: mackerel-plugin-mysql-lite

mackerel-plugin-mysql-lite: mackerel-plugin-mysql-lite.go
	go build $(LDFLAGS) -o mackerel-plugin-mysql-lite

linux: mackerel-plugin-mysql-lite.go
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o mackerel-plugin-mysql-lite

clean:
	rm -rf mackerel-plugin-mysql-lite

check:
	go test ./...

tag:
	git tag v${VERSION}
	git push origin v${VERSION}
	git push origin master
