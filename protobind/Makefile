PROTOBIND = protobind
PROTOBIND_BIN = $(shell which $(PROTOBIND))

.PHONY: install
install:
	@go install

.PHONY: clean
clean:
	$(if $(PROTOBIND_BIN), $(shell rm $(PROTOBIND_BIN)))
