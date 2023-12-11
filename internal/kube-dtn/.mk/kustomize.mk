.PHONY: cni-manifests
## Build CNI plugin configuration
cni-manifests: manifests kustomize
	cd config/cni && $(KUSTOMIZE) edit set image ${CNI_IMG}:${COMMIT}
	$(KUSTOMIZE) build config/cni > cni.yaml

.PHONY: cni-deploy
## Deploy CNI plugin into cluster
cni-deploy: manifests kustomize
	cd config/cni && $(KUSTOMIZE) edit set image ${CNI_IMG}:${COMMIT}
	$(KUSTOMIZE) build config/cni | kubectl apply -f -

.PHONY: cni-undeploy
## Undeploy CNI plugin from cluster
cni-undeploy:
	$(KUSTOMIZE) build config/cni | kubectl delete --ignore-not-found=$(ignore-not-found) -f -
