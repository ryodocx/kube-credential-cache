### Requirements
- docker
- kind
- kubectl
- jq
- yj
- gdat
- gsed
- aws

### Usage

```sh
export KUBECONFIG="kubeconfig.yaml"
export KUBE_CREDENTIAL_CACHE_FILE=".tmp/cache.json"
export KUBECONFIG_EKS=<your kubeconfig to access EKS>
export KUBECONTEXT_EKS=<your kube-context to access EKS>
export AWS_PROFILE_EKS=<your AWS_PROFILE to access EKS>

# run benchmark
make bench

# teadown
make reset
```

### Result


raw data

```sh
$ go test -bench . -cpu 1 -benchtime 10s
goos: darwin
goarch: arm64
pkg: github.com/ryodocx/kube-credential-cache/benchmark
BenchmarkKubectlFast                  64         158413617 ns/op # kubectl version --user kind-kcc-bench (no wait)
BenchmarkKubectlCache                 68         160467031 ns/op # kubectl version --user cache (use get-credential-wait.sh with kcc-cache)
BenchmarkKubectlSlow                  16         701907578 ns/op # kubectl version --user slow  (use get-credential-wait.sh only)
BenchmarkGetTokenCache              6841           1739251 ns/op # kcc-cache sh get-credential-wait.sh
BenchmarkGetTokenOriginal             21         532414587 ns/op # sh get-credential-wait.sh
BenchmarkGetTokenEKS                  21         530606196 ns/op # aws eks get-token --cluster-name example
PASS
ok      github.com/ryodocx/kube-credential-cache/benchmark      84.108s
```


![](./graph_kubectl.svg)

![](./graph_credential.svg)