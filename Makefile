.PHONY: all deploy health-check monitoring stop-monitoring clean

# Main command to run everything
all: deploy monitoring

# Deploy Ethereum node (includes health check)
deploy:
	@echo "=== Deploying Ethereum Node ==="
	./scripts/deploy-ethereum.sh

# Run health check separately when needed
health-check:
	@echo "=== Running Health Check ==="
	./scripts/run-health-check.sh

# Forward only monitoring services (Grafana, Prometheus)
monitoring:
	@echo "=== Port forwarding monitoring services ==="
	@kubectl port-forward svc/ethereum-grafana 3000:80 > /dev/null 2>&1 & echo "$$!" > .grafana-pid; \
	kubectl port-forward svc/ethereum-prometheus-server 9090:80 > /dev/null 2>&1 & echo "$$!" > .prometheus-pid; \
	echo "Monitoring services port forwarding started:"; \
	echo "- Grafana: http://localhost:3000 (admin/admin123)"; \
	echo "- Prometheus: http://localhost:9090"; \
	echo "Use 'make stop-monitoring' to stop port forwarding"

# Stop monitoring port forwarding
stop-monitoring:
	@echo "=== Stopping monitoring port forwarding ==="
	@[ -f .grafana-pid ] && kill $$(cat .grafana-pid) 2>/dev/null && rm .grafana-pid || true
	@[ -f .prometheus-pid ] && kill $$(cat .prometheus-pid) 2>/dev/null && rm .prometheus-pid || true
	@echo "Monitoring port forwarding stopped"

# Clean up
clean: stop-monitoring
	@echo "=== Cleaning Up ==="
	@rm -f scripts/eth-health-check
	@rm -f .grafana-pid .prometheus-pid