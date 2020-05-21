.PHONY: all
all: clean build-linux-amd64 build-darwin-amd64

.PHONY: build-linux-amd64
build-linux-amd64:
	mkdir -p out/linux-amd64
	env GOOS=linux GOARCH=amd64 go build -o ./out/linux-amd64/gmail-cleaner .

.PHONY: build-darwin-amd64
build-darwin-amd64:
	mkdir -p out/darwin-amd64
	env GOOS=darwin GOARCH=amd64 go build -o ./out/darwin-amd64/gmail-cleaner .

.PHONY: clean
clean:
	rm -rf ./out

.PHONY: deploy-burr
deploy-burr: clean build-linux-amd64
	scp ./out/gmail-cleaner burr:~/gmail-cleaner
