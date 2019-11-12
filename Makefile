.PHONY: all
all: build

.PHONY: build
build: clean
	env GOOS=linux GOARCH=amd64 go build -o ./out/gmail-cleaner .

.PHONY: clean
clean:
	rm -rf ./out

.PHONY: deploy-burr
deploy-burr: clean build
	scp ./out/gmail-cleaner burr:~/gmail-cleaner