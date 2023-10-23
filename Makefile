lint:
	@docker run     \
    --rm            \
    -v $(CURDIR):/app  \
    -w /app         \
    golangci/golangci-lint:v1.53-alpine golangci-lint run -v