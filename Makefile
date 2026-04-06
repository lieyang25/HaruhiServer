APP_NAME=haruhiserver
CMD_DIR=./cmd/haruhi_server

.PHONY: run build tidy clean

run:
	go run $(CMD_DIR)

build:
	go build -o bin/$(APP_NAME) $(CMD_DIR)

tidy:
	go mod tidy

clean:
	rm -rf bin