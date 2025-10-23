export DOCKER_REGISTRY_URL ?= sanshunoisky
export IMAGE_NAME ?= erp-pp
export IMAGE_TAG ?= $(CLUSTER_NAME).latest

IMG ?= ${DOCKER_REGISTRY_URL}/${IMAGE_NAME}:${IMAGE_TAG}

GO_PACKAGES = ./internal/...

.PHONY: test

template:
	helm template service-broker ./charts/service-broker \
	--values ./charts/service-broker/values.yaml \
	--output-dir ../service-broker/manifests \
	--namespace service-broker

docker-build:
	docker build . -t ${IMG}

docker-push: docker-build
	docker push ${IMG}

check-cluster-name:
	@{ \
	if [ -z "$(CLUSTER_NAME)" ]; then \
		echo "Environment variable CLUSTER_NAME is mandatory"; \
		exit 1; \
	else \
		echo "\nCLUSTER_NAME variable set to $(CLUSTER_NAME)"; \
	fi \
	}

test-cover:
	go test -v $(GO_PACKAGES) -coverprofile cover.out -coverpkg=$(GO_PACKAGES) -covermode=count
	go tool cover -html=cover.out


test:
	go test $(GO_PACKAGES) -coverprofile cover.out -coverpkg=$(GO_PACKAGES) -mod=vendor