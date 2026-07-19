clean:
	rm -f coverage.out

SUPRESS_OUTPUT = > /dev/null

test-coverage:
	@go test -coverprofile=coverage.out ./... $(SUPRESS_OUTPUT)
	@go tool cover -html=coverage.out $(SUPRESS_OUTPUT)
	@make clean

test-total-coverage:
	@go test ./... -coverprofile=coverage.out $(SUPRESS_OUTPUT)
	@go tool cover -func=coverage.out | grep total | awk '{print $$1, $$3}'
	@make clean $(SUPRESS_OUTPUT)
