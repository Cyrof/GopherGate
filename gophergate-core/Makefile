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
