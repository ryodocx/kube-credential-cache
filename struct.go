package main

import (
	"time"
)

// https://kubernetes.io/docs/reference/config-api/client-authentication.v1
type ClientAuthentication struct {
	APIVersion string   `json:"apiVersion"`
	Kind       string   `json:"kind"`
	Spec       struct{} `json:"spec"`
	Status     struct {
		ExpirationTimestamp time.Time `json:"expirationTimestamp"`
		Token               string    `json:"token"`
	} `json:"status"`
}

type CacheFile struct {
	Tokens map[string]ClientAuthentication `json:"tokens"`
}
