#!/bin/sh

sleep 0.23
cat <<EOS
{
  "kind": "ExecCredential",
  "apiVersion": "client.authentication.k8s.io/v1beta1",
  "spec": {},
  "status": {
    "expirationTimestamp": "$(gdate +%Y-%m-%dT%H:%M:%SZ --utc --date '15 min')",
    "clientCertificateData": "$(cat kubeconfig.yaml | yj | jq -r '.users[] | select(.name == "kind-kcc-bench") | .user["client-certificate-data"]' | base64 -D | gsed -z 's/\n/\\n/g')",
    "clientKeyData": "$(cat kubeconfig.yaml | yj | jq -r '.users[] | select(.name == "kind-kcc-bench") | .user["client-key-data"]' | base64 -D | gsed -z 's/\n/\\n/g')"
  }
}
EOS
