package main

import (
	"log"
	"os"
	"time"
)

var (
	cacheFile     string        = "~/.kube/cache/token.json"
	refreshMargin time.Duration = time.Second * 30
)

func init() {
	if e := os.Getenv("EKS_TOKEN_CACHE_FILE"); e != "" {
		cacheFile = e
	}

	if e := os.Getenv("EKS_TOKEN_CACHE_REFRESH_MARGIN"); e != "" {
		d, err := time.ParseDuration(e)
		if err != nil {
			log.Fatalf("invalid environment variable 'EKS_TOKEN_CACHE_REFRESH_MARGIN': %s", err.Error())
		}
		refreshMargin = d
	}
}
