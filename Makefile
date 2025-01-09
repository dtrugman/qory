OUTPUT_DIR=$(CURDIR)/dist

BINARY_NAME=qory

build: format
	goreleaser build \
		--snapshot \
		--clean \
		--single-target \
		--output $(OUTPUT_DIR)/$(BINARY_NAME)

release: format
	goreleaser release \
		--clean

format:
	go fmt ./...
