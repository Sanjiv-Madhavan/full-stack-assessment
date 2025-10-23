export DOCKER_REGISTRY_URL ?= sanshunoisky
export IMAGE_NAME ?= full-stack-backend
export IMAGE_TAG ?= $(CLUSTER_NAME).latest

IMG ?= ${DOCKER_REGISTRY_URL}/${IMAGE_NAME}:${IMAGE_TAG}

GO_PACKAGES = ./internal/...

.PHONY: test

template:
	helm template full-stack-backend ./charts/full-stack-backend \
	--values ./charts/full-stack-backend/values.yaml \
	--output-dir ./charts/full-stack-backend/manifests \
	--namespace full-stack-backend

docker-build:
	docker build . -t ${IMG}

docker-push: docker-build
	docker push ${IMG}

apply:
	kubectl apply --recursive --filename ./charts/full-stack-backend/manifests/

delete:
	kubectl delete --recursive --filename ./charts/full-stack-backend/manifests/

check-cluster-name:
	@{ \
	if [ -z "$(CLUSTER_NAME)" ]; then \
		echo "Environment variable CLUSTER_NAME is mandatory"; \
		exit 1; \
	else \
		echo "\nCLUSTER_NAME variable set to $(CLUSTER_NAME)"; \
	fi \
	}

create-namespace:
	-kubectl create namespace full-stack-backend

generate-values:
	@echo "Generating values.yaml with envsubst"
	cd charts && envsubst < full-stack-backend/values-template.yaml > full-stack-backend/values.yaml


test-cover:
	go test -v $(GO_PACKAGES) -coverprofile cover.out -coverpkg=$(GO_PACKAGES) -covermode=count
	go tool cover -html=cover.out


test:
	go test $(GO_PACKAGES) -coverprofile cover.out -coverpkg=$(GO_PACKAGES) -mod=vendor


deploy-to-cluster: check-cluster-name generate-values docker-push template create-namespace apply
