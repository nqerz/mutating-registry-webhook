.PHONY: build test deploy helm-package test-deploy verify-deployment

IMAGE_NAME := registry-mutation-webhook
IMAGE_TAG := latest
REGISTRY := nqerz
CHART_VERSION := 0.1.0
build:
	go build -o bin/webhook cmd/main.go
	chmod +x bin/webhook

docker-build:
	docker build -t $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG) .

docker-push:
	docker push $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)

test:
	go test ./...

deploy: docker-build docker-push
	helm upgrade --install registry-webhook ./charts

generate-certs:
	chmod +x ./scripts/tls-gen.sh
	./scripts/tls-gen.sh

helm-package:
	helm package ./charts --version $(CHART_VERSION) --app-version $(IMAGE_TAG)

dry-run:
	helm install --dry-run --debug registry-webhook ./charts

test-deploy: helm-package
	kubectl config use-context docker-desktop
	kubectl apply -f secret.yaml
	helm upgrade --install --wait registry-webhook ./registry-mutation-webhook-*.tgz

# Verify the deployment
verify-deployment:
	kubectl get pods -l app=mutating-registry-webhook
	kubectl get mutatingwebhookconfigurations.admissionregistration.k8s.io registry-webhook

# Clean up build artifacts
clean:
	rm -rf bin/ *.tgz