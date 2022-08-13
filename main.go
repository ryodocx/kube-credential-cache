package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

// https://kubernetes.io/docs/reference/config-api/client-authentication.v1
type ClientAuthentication struct {
	APIVersion string   `json:"apiVersion"`
	Kind       string   `json:"kind"`
	Spec       struct{} `json:"spec"`
	Status     struct {
		ExpirationTimestamp   time.Time `json:"expirationTimestamp"`
		Token                 string    `json:"token,omitempty"`
		ClientCertificateData string    `json:"clientCertificateData,omitempty"`
		ClientKeyData         string    `json:"clientKeyData,omitempty"`
	} `json:"status"`
}

type CacheFile struct {
	Tokens map[string]ClientAuthentication `json:"tokens"`
}

func main() {
	// configuration
	var (
		cacheFilepath string
		refreshMargin time.Duration = time.Second * 30
	)
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

	// open file
	f, err := os.OpenFile(cacheFilepath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(path.Dir(cacheFilepath), 0700); err != nil {
				log.Fatalf("mkdir failed: %s", err)
			}
			f, err = os.OpenFile(cacheFilepath, os.O_RDWR|os.O_CREATE, 0600)
			if err != nil {
				log.Fatalf("file open failed(after mkdir): %s", err)
			}
		} else {
			log.Fatalf("file open failed: %s", err)
		}
	}
	defer f.Close()

	// read file
	updated := false
	cacheFile := CacheFile{}
	bytes, err := io.ReadAll(f)
	if err != nil {
		log.Fatalf("file read failed: %s", err)
	}
	if len(bytes) > 0 {
		if err := json.Unmarshal(bytes, &cacheFile); err != nil {
			log.Fatalf("json.Unmarshal() failed(read cache file): %s", err)
		}
	}
	defer func() {
		// update cache file
		if updated {
			qpanic := func(err error) {
				if err != nil {
					panic(err)
				}
			}

			err := f.Truncate(0)
			qpanic(err)
			bytes, err := json.Marshal(cacheFile)
			qpanic(err)
			_, err = f.WriteAt(bytes, 0)
			qpanic(err)
		}
	}()

	// check cache
	if len(cacheFile.Tokens) == 0 {
		cacheFile.Tokens = map[string]ClientAuthentication{}
	}
	cache, ok := cacheFile.Tokens[mapKey()]
	if !ok || ok && time.Until(cache.Status.ExpirationTimestamp) < refreshMargin {
		// refresh
		tmpCache := ClientAuthentication{}

		if len(os.Args) < 2 {
			log.Fatalf("not enough command at args")
		}
		cmd := exec.Command(os.Args[1], os.Args[2:]...)
		cmd.Stderr = os.Stderr
		cmd.Env = os.Environ()
		bytes, err := cmd.Output()

		if err != nil {
			log.Fatalf("read command output failed: %s\noutput: %s", err, string(bytes))
		}

		if err := json.Unmarshal(bytes, &tmpCache); err != nil {
			log.Fatalf("json.Unmarshal() failed(read command output): %s\nactual stdout: %s", err, string(bytes))
		}

		if time.Until(tmpCache.Status.ExpirationTimestamp) < refreshMargin {
			log.Fatalf("Obtained token has expired: %s", string(bytes))
		}

		cacheFile.Tokens[mapKey()] = tmpCache
		updated = true
	}

	// print
	output, err := json.Marshal(cacheFile.Tokens[mapKey()])
	if err != nil {
		log.Fatalf("json.Marshal() failed: %s", err)
	}
	fmt.Println(string(output))
}

func mapKey() string {
	return strings.Join(os.Args[1:], " ")
}
