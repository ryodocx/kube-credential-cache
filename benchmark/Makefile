export KUBECONFIG = kubeconfig.yaml
export KUBE_CREDENTIAL_CACHE_FILE := .tmp/cache.json

bench: setup
	@rm -f .tmp/cache.json
	kubectl version --short 2>/dev/null
	go test -bench . -cpu 1 -benchtime 10s

setup: .tmp/check ${KUBECONFIG} kcc-cache

reset:
	kind delete cluster --name kcc-bench || true
	rm ${KUBECONFIG} || true
	rm -rf .tmp
	rm -f kcc-cache

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

${KUBECONFIG}: kind.yaml
	kind create cluster --config kind.yaml
	cat additional_kubeconfig.yaml >> ${KUBECONFIG}

kcc-cache:
	go build -o $@ ../cmd/kcc-cache/
