## Architecture
![](./summary.drawio.svg)

## Sequence

### Without cache

```mermaid
sequenceDiagram
    actor user as User
    participant tool as kubectl
    participant client as K8s client library
    participant plugin as credential plugin
    participant k8s as Kubernetes

    user ->>+  tool: run
    tool ->>+  client: init
    client ->> client: load kubeconfig
    tool ->> client: method call
    rect rgb(200, 150, 255)
      client ->>+ plugin: request credential
      plugin ->> plugin: slow process
      plugin -->>- client: credential
    end
    loop
      client ->> k8s: API call
      k8s -->> client: result
      client ->> client: process
    end
    client -->> tool: result
    tool -->> user: result

    deactivate client
    deactivate tool
```

### With cache

```mermaid
sequenceDiagram
    actor user as User
    participant tool as kubectl
    participant client as K8s client library
    participant cache as [kcc-cache]
    participant plugin as credential plugin
    participant k8s as Kubernetes

    user ->>+  tool: run
    tool ->>+  client: init
    client ->> client: load kubeconfig
    tool ->> client: method call
    client ->>+ cache: request credential
    cache ->> cache: check cache
    alt cache not exists or expired
      rect rgb(200, 150, 255)
        cache ->>+ plugin: request credential
        plugin ->> plugin: slow process
        plugin -->>- cache: credential
      end
    end
    cache -->>- client: credential
    loop
      client ->> k8s: API call
      k8s -->> client: result
      client ->> client: process
    end
    client -->> tool: result
    tool -->> user: result

    deactivate client
    deactivate tool
```
