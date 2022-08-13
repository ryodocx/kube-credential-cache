export KUBECONFIG = kubeconfig.yaml
export KUBE_CREDENTIAL_CACHE_FILE := .tmp/cache.json

default: .tmp/check kubeconfig.yaml
	@echo completed

reset:
	kind delete cluster --name kcc-bench || true
	rm ${KUBECONFIG} || true
	rm -rf .tmp
	make default

.tmp/check: .tmp/mkdir
	docker version --format '{{.Server.Version}}'
	kind version -q
	kubectl version --client --short
	jq --version
	yj -v
	gdate
	gsed --version
	aws --version
	touch $@

.tmp/mkdir:
	mkdir -p .tmp
	touch $@

kind-create: ${KUBECONFIG}
${KUBECONFIG}: kind.yaml
	kind create cluster --config kind.yaml
	cat additional_kubeconfig.yaml >> ${KUBECONFIG}

kcc-cache:
	go build -o $@ ../cmd/kcc-cache/

bench: default kcc-cache
	rm -f .tmp/cache.json
	go test -bench . -cpu 1 -benchtime 10s
