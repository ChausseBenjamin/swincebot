BUILD_DIR=./build
APP=swincebot

all: codegen clean compile

codegen:
	sqlc generate

compile: codegen
	mkdir -p $(BUILD_DIR)
	go run ./internal/autogen > $(BUILD_DIR)/$(APP).1
	go build -o $(BUILD_DIR)/$(APP) .

clean:
	rm -rf $(BUILD_DIR)

.PHONY: run
run: codegen
	./resources/local_dev.sh

