.PHONY: dev
dev:
	@if [ -z "$(MIXIN)" ]; then \
		echo "Usage: make dev MIXIN=<mixin>"; \
		echo "Example: make dev MIXIN=otel-app-semantic"; \
		exit 1; \
	fi
	grafanactl resources serve --script 'go run $(MIXIN)/main.go' --watch './$(MIXIN)' --watch './common'