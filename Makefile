.PHONY: all
all:
	env GOOS=linux GOARCH=amd64 go build -o ./out/gmail-cleaner .

.PHONY: clean
clean:
	rm -rf ./out