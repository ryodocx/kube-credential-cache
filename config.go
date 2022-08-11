package main

import (
	"log"
	"os"
	"path"
	"time"
)

var (
	cacheFilepath string
	refreshMargin time.Duration = time.Second * 30
)

func init() {
	if e := os.Getenv("K8S_TOKEN_CACHE_FILE"); e != "" {
		cacheFilepath = e
	} else {
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			log.Fatalf("can't find CacheDir. fix error or set 'K8S_TOKEN_CACHE_FILE': %s", err)
		}
		cacheFilepath = path.Join(cacheDir, "k8s-token-cache/token.json")
		// mac: /Users/${USER}/Library/Caches/k8s-token-cache/token.json
	}

	if e := os.Getenv("K8S_TOKEN_CACHE_REFRESH_MARGIN"); e != "" {
		d, err := time.ParseDuration(e)
		if err != nil {
			log.Fatalf("invalid environment variable 'K8S_TOKEN_CACHE_REFRESH_MARGIN': %s", err.Error())
		}
		refreshMargin = d
	}
}
