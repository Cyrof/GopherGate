CLUSTER_NAME := gophergate-dev
CONFIG_FILE := k3d-config.yaml

.PHONY: create delete restart status

create:
	@echo "Creating k3d cluster: $(CLUSTER_NAME)"
	k3d cluster create --config $(CONFIG_FILE)

delete:
	@echo "Deleting k3d cluster: $(CLUSTER_NAME)"
	k3d cluster delete $(CLUSTER_NAME)

restart: delete create

status:
	@echo "k3d clusters:"
	k3d cluster list
	@echo "\nkubectl nodes:"
	- kubectl get nodes -o wide || echo "Cluster not running"