### Requirements
- docker
- kind
- kubectl
- jq
- yj
- gdat
- gsed
- aws

### Run

```sh
make bench
```

### Result


raw data

```sh
$ go test -bench . -cpu 1 -benchtime 10s
goos: darwin
goarch: arm64
pkg: github.com/ryodocx/kube-credential-cache/benchmark
BenchmarkKubectlFast                  64         158413617 ns/op # kubectl version --user kind-kcc-bench (no wait)
BenchmarkKubectlCache                 68         160467031 ns/op # kubectl version --user cache (use get-token-wait.sh with kcc-cache)
BenchmarkKubectlSlow                  16         701907578 ns/op # kubectl version --user slow  (use get-token-wait.sh only)
BenchmarkGetTokenCache              6841           1739251 ns/op # kcc-cache sh get-token-wait.sh
BenchmarkGetTokenOriginal             21         532414587 ns/op # sh get-token-wait.sh
BenchmarkGetTokenEKS                  21         530606196 ns/op # aws eks get-token --cluster-name example
PASS
ok      github.com/ryodocx/kube-credential-cache/benchmark      84.108s
```


![](./graph_kubectl.svg)

![](./graph_credential.svg)
