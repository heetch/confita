VAULT_ADDR ?= http://127.0.0.1:8200

.PHONY: test

test:
	go build
	VAULT_ADDR="$(VAULT_ADDR)" \
	go test -v -race -cover -timeout=1m ./... && echo OK || (echo FAIL && exit 1)
