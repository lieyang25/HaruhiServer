APP_NAME=haruhiserver
CMD_DIR=./cmd/haruhiserver

.PHONY: run build test tidy clean

run:
	go run $(CMD_DIR)

build:
	go build -o bin/$(APP_NAME) $(CMD_DIR)

test:
	go test ./test/...

tidy:
	go mod tidy

clean:
	rm -rf bin
