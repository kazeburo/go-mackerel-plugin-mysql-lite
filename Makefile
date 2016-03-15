VERSION=0.0.1

all: mackerel-plugin-mysql-lite

.PHONY: mackerel-plugin-mysql-lite

gom:
	go get -u github.com/mattn/gom

bundle:
	gom install

mackerel-plugin-mysql-lite: mackerel-plugin-mysql-lite.go
	gom build -o mackerel-plugin-mysql-lite

linux: mackerel-plugin-mysql-lite.go
	GOOS=linux GOARCH=amd64 gom build -o mackerel-plugin-mysql-lite

fmt:
	go fmt ./...

dist:
	git archive --format tgz HEAD -o mackerel-plugin-mysql-lite-$(VERSION).tar.gz --prefix mackerel-plugin-mysql-lite-$(VERSION)/

clean:
	rm -rf mackerel-plugin-mysql-lite mackerel-plugin-mysql-lite-*.tar.gz

