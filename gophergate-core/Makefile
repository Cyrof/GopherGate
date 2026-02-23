PROTO_DIR := api/proto
OUT_DIR := pkg/gen

.PHONY: test itest

test:
	@echo "[*] Running tests for gophergate-core..."
	@echo ""
	go test -race -covermode=atomic -coverpkg=./... ./test -count=1 -v
	@echo ""
	@echo "[+] All tests completed!"

itest: 
	@echo "[*] Spinning up Postgres and running dbx integration tests..."
	GOFLAGS="-tags=integration" go test ./test -run TestDBX_DockertestIntegration -v -count=1 -timeout=8m
	@echo "[+] Integration tests done."

proto:
	@echo "Generating gRPC stubs..."
	protoc \
		--proto_path=$(PROTO_DIR) \
		--go_out=$(OUT_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(OUT_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/gateway/v1/*.proto