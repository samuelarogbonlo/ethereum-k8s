.PHONY: all deploy build-health-check health-check clean
# Main command to run everything
all: deploy build-health-check health-check
# Deploy Ethereum node
deploy:
	@echo "=== Deploying Ethereum Node ==="
	./deploy-ethereum.sh
# Build health check tool
build-health-check:
	@echo "=== Building Health Check Tool ==="
	cd health-check && go build -o eth-health-check eth-health-check.go
# Run health check
health-check: build-health-check
	@echo "=== Running Health Check ==="
	@kubectl port-forward svc/geth 8545:8545 > /dev/null 2>&1 & PID1=$$!; \
	kubectl port-forward svc/lighthouse 5052:5052 > /dev/null 2>&1 & PID2=$$!; \
	echo "Port forwarding started, waiting 3 seconds..."; \
	sleep 3; \
	cd health-check && ./eth-health-check; \
	EXIT_CODE=$$?; \
	echo "Cleaning up port forwarding..."; \
	kill $$PID1 $$PID2 2>/dev/null || true; \
	exit $$EXIT_CODE
# Clean up
clean:
	@echo "=== Cleaning Up ==="
	@rm -f health-check/eth-health-check
# Help
help:
	@echo "Available targets:"
	@echo "  make              - Deploy Ethereum node and run health check"
	@echo "  make deploy       - Deploy Ethereum node only"
	@echo "  make health-check - Run health check only"
	@echo "  make clean        - Clean up build artifacts"
	@echo "  make help         - Show this help"