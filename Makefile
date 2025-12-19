CLUSTER_NAME := gophergate-dev
CONFIG_FILE := k3d-config.yaml
K3D_CONTEXT  := k3d-$(CLUSTER_NAME)
PROD_CONTEXT ?= kpi

.PHONY: create delete restart status use-k3d use-prod ensure-k3d-context

create:
	@echo "Creating k3d cluster: $(CLUSTER_NAME)"
	k3d cluster create --config $(CONFIG_FILE)
	@$(MAKE) use-k3d

delete:
	@echo "Switching kubectl context back to prod: $(PROD_CONTEXT)"
	@if kubectl config  get-contexts -o name | grep -qx "$(PROD_CONTEXT)"; then \
		kubectl config use-context "$(PROD_CONTEXT)"; \
	else \
		echo "WARNING: PROD_CONTEXT '$(PROD_CONTEXT)' not found. kubectl left unchanged."; \
	fi
	@echo "Deleting k3d cluster: $(CLUSTER_NAME)"
	k3d cluster delete $(CLUSTER_NAME)

restart: delete create

use-k3d:
	@if kubectl config get-contexts -o name | grep -qx "$(K3D_CONTEXT)"; then \
		echo "Switching kubectl context -> $(K3D_CONTEXT)"; \
		kubectl config use-context "$(K3D_CONTEXT)"; \
	else \
		echo "k3d context $(K3D_CONTEXT) not found (cluster not created yet)."; \
		echo "Run: make create"; \
		exit 1; \
	fi

use-prod:
	@echo "Switching kubectl context -> $(PROD_CONTEXT)"
	kubectl config use-context "$(PROD_CONTEXT)"

status:
	@echo "k3d clusters:"
	k3d cluster list
	@echo "\nkubectl current-context:"
	- kubectl config current-context || true
	@echo "\nkubectl nodes:"
	- kubectl get nodes -o wide || echo "Cluster not running"
