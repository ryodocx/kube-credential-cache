# kcc: kube-credential-cache
Faster access to kubernetes!
especially, for kubectl + EKS

## Features
Work as caching proxy of [ExecCredential](https://kubernetes.io/docs/reference/config-api/client-authentication.v1beta1/#client-authentication-k8s-io-v1beta1-ExecCredential) object, when use [client-go credential plugins](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#client-go-credential-plugins) of Kubernetes. (e.g. kubectl)

- Caching
  - [x] Cache [ExecCredential](https://kubernetes.io/docs/reference/config-api/client-authentication.v1beta1/#client-authentication-k8s-io-v1beta1-ExecCredential) object
  - [ ] Async credential refresh
- Cache file
  - [ ] Encryption
- kubeconfig
  - [ ] kubeconfig optimizer (inject cache command automatically)

## Effects
A one of notable effect is, when used [`aws eks update-kubeconfig`](https://docs.aws.amazon.com/eks/latest/userguide/create-kubeconfig.html) to access EKS. about 500ms faster!

<!-- TODO: graph -->

## Installation

```sh
go install github.com/ryodocx/kube-credential-cache@latest
```

## Usage

* edit your kubeconfig
  * set `kcc-cache` to command
  * original command move to args
  * **remove env** because `kcc-cache` only uses args for cache key

EKS

```diff
kind: Config
apiVersion: v1
clusters: [...]
contexts: [...]
current-context: <your-current-context>
preferences: {}
users:
  - name: user-name
    user:
      exec:
        apiVersion: client.authentication.k8s.io/v1beta1
-       command: aws
+       command: kcc-cache
        args:
+         - aws
          - --region
          - <your-region>
          - eks
          - get-token
          - --cluster-name
          - <your-cluster>
+         - --profile
+         - <your-profile>
-       env:
-         - name: AWS_PROFILE
-           value: <your-profile>
```

EKS with [aws-vault](https://github.com/99designs/aws-vault)

```diff
kind: Config
apiVersion: v1
clusters: [...]
contexts: [...]
current-context: <your-current-context>
preferences: {}
users:
  - name: user-name
    user:
      exec:
        apiVersion: client.authentication.k8s.io/v1beta1
-       command: aws
+       command: kcc-cache
        args:
+         - aws-vault
+         - exec
+         - <your-profile>
+         - --
+         - aws
          - --region
          - <your-region>
          - eks
          - get-token
          - --cluster-name
          - <your-cluster>
-       env:
-         - name: AWS_PROFILE
-           value: <your-profile>
```

kubeconfig specification
* https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/
* https://github.com/kubernetes/client-go/blob/release-1.24/tools/clientcmd/api/v1/types.go#L28-L52

## Configration

| Environment variable                 | default                                                                                                                                                                                                                                        | description                  |
|--------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|------------------------------|
| KUBE_CREDENTIAL_CACHE_FILE           | macOS:</br>`~/Library/Caches/kube-credential-cache/cache.json`</br>Linux:</br>`$XDG_CACHE_HOME/kube-credential-cache/cache.json`</br>`~/.cache/kube-credential-cache/cache.json`</br>Windows:</br>`%AppData%\kube-credential-cache\cache.json` | path of Cache file           |
| KUBE_CREDENTIAL_CACHE_REFRESH_MARGIN | `30s`                                                                                                                                                                                                                                          | margin of credential refresh |
