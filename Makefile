.PHONY: build test docker-build helm-lint deploy clean help

SERVICES = api-gateway account-service auth-service notification-service \
           credit-service compliance-service data-product-service config-service \
           payment-service

REGISTRY ?= ghcr.io/your-org/account-center
TAG ?= latest
NAMESPACE ?= account-center
RELEASE ?= account-center

help:
	@echo "Account Center V2.0 - Makefile Commands"
	@echo ""
	@echo "  make build          Build all Go services"
	@echo "  make test           Test all Go services"
	@echo "  make docker-build   Build all Docker images"
	@echo "  make docker-push    Push all Docker images"
	@echo "  make helm-lint      Lint the Helm chart"
	@echo "  make deploy         Deploy with Helm"
	@echo "  make clean          Clean build artifacts"

build:
	@for svc in $(SERVICES); do \
		echo "==> Building $$svc"; \
		cd $$svc && go build -o ../bin/$$svc ./cmd/main.go && cd ..; \
	done

test:
	@for svc in $(SERVICES); do \
		echo "==> Testing $$svc"; \
		cd $$svc && go test -v -race ./... && cd ..; \
	done

docker-build:
	@for svc in $(SERVICES); do \
		echo "==> Docker building $$svc"; \
		if [ "$$svc" = "config-service" ]; then \
			docker build -t $(REGISTRY)/$$svc:$(TAG) -f $$svc/Dockerfile .; \
		else \
			docker build -t $(REGISTRY)/$$svc:$(TAG) $$svc/; \
		fi; \
	done
	@echo "==> Docker building web-ui"
	@cd web-ui && npm ci && npm run build && cd ..
	@docker build -t $(REGISTRY)/web-ui:$(TAG) web-ui/

docker-push:
	@for svc in $(SERVICES) web-ui; do \
		echo "==> Pushing $$svc"; \
		docker push $(REGISTRY)/$$svc:$(TAG); \
	done

helm-lint:
	helm lint helm/account-center --strict
	helm template account-center helm/account-center > /dev/null

deploy: helm-lint
	helm upgrade --install $(RELEASE) ./helm/account-center \
		--namespace $(NAMESPACE) \
		--create-namespace \
		--wait \
		--timeout 5m \
		--set global.imageRegistry=$(REGISTRY) \
		--set tag=$(TAG)

clean:
	rm -rf bin/

.PHONY: perf-test perf-smoke perf-load perf-stress

perf-test: perf-smoke perf-load perf-stress

perf-smoke:
	k6 run tests/perf/smoke.js

perf-load:
	k6 run tests/perf/load.js

perf-stress:
	k6 run tests/perf/stress.js
