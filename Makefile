OUTPUT_DIR=$(CURDIR)/dist

BINARY_NAME=qory

build:
	goreleaser build --snapshot --clean --single-target --output $(OUTPUT_DIR)/$(BINARY_NAME)
