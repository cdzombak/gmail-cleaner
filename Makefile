.PHONY: all
all: build-linux-amd64

.PHONY: build-linux-amd64
build-linux-amd64: clean
	mkdir -p out
	env GOOS=linux GOARCH=amd64 go build -o ./out/gmail-cleaner .

.PHONY: clean
clean:
	rm -rf ./out

.PHONY: deploy-burr
deploy-burr: clean build-linux-amd64
	scp ./out/gmail-cleaner burr:~/gmail-cleaner
