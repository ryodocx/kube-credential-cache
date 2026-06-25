package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/gofrs/flock"
	"github.com/ryodocx/kube-credential-cache/internal/util"
)

type CacheFile struct {
	Credentials map[string]ClientAuthentication `json:"credentials"`
}

// https://kubernetes.io/docs/reference/config-api/client-authentication.v1
type ClientAuthentication struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Status     struct {
		ExpirationTimestamp   time.Time `json:"expirationTimestamp"`
		Token                 string    `json:"token,omitempty"`
		ClientCertificateData string    `json:"clientCertificateData,omitempty"`
		ClientKeyData         string    `json:"clientKeyData,omitempty"`
	} `json:"status"`
}

func main() {
	// configuration
	var (
		cacheFilepath   string
		refreshMargin   time.Duration = time.Second * 30
		cacheKeyEnvlist []string      = []string{"KUBE_CREDENTIAL_CACHE_USER", "AWS_PROFILE", "AWS_REGION", "AWS_VAULT"}
	)
	if e := os.Getenv("KUBE_CREDENTIAL_CACHE_FILE"); e != "" {
		cacheFilepath = e
	} else {
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			util.Fatal(nil, "can't find CacheDir. fix error or set 'KUBE_CREDENTIAL_CACHE_FILE': %s", err)
		}
		cacheFilepath = path.Join(cacheDir, "kube-credential-cache", "cache.json")
	}
	if e := os.Getenv("KUBE_CREDENTIAL_CACHE_REFRESH_MARGIN"); e != "" {
		d, err := time.ParseDuration(e)
		if err != nil {
			util.Fatal(nil, "invalid environment variable 'KUBE_CREDENTIAL_CACHE_REFRESH_MARGIN': %s", err.Error())
		}
		refreshMargin = d
	}
	if e := os.Getenv("KUBE_CREDENTIAL_CACHE_CACHEKEY_ENV_LIST"); e != "" {
		cacheKeyEnvlist = strings.Split(e, ",")
	}

	// cache key
	var cacheKey string = strings.Join(os.Args[1:], " ")
	{
		env := ""
		for _, key := range cacheKeyEnvlist {
			v := os.Getenv(key)
			if v == "" {
				continue
			}
			env = fmt.Sprintf("%s %s='%s'", env, key, v)
		}
		if env != "" {
			cacheKey = fmt.Sprintf("%s # env:%s", cacheKey, env)
		}
	}

	// ensure directory exists
	if err := os.MkdirAll(path.Dir(cacheFilepath), 0700); err != nil {
		util.Fatal(nil, "mkdir failed: %s", err)
	}

	// first read (lock free)
	cacheFile := CacheFile{}
	bytes, err := os.ReadFile(cacheFilepath)
	if err != nil && !os.IsNotExist(err) {
		util.Fatal(nil, "file read failed: %s", err)
	}
	if len(bytes) > 0 {
		if err := json.Unmarshal(bytes, &cacheFile); err != nil {
			util.Log("json.Unmarshal() failed(read cache file): %s\n...Corruption detected, recreate cache file", err)
		}
	}

	if len(cacheFile.Credentials) == 0 {
		cacheFile.Credentials = map[string]ClientAuthentication{}
	}

	cache, ok := cacheFile.Credentials[cacheKey]
	updated := false

	// check if credential needs refreshing
	if !ok || ok && time.Until(cache.Status.ExpirationTimestamp) < refreshMargin {
		// refresh
		tmpCache := ClientAuthentication{}

		if len(os.Args) < 2 {
			util.Fatal(nil, "not enough command at args")
		}

		// Run external command (lock free, no blocking for other cache keys)
		cmd := exec.Command(os.Args[1], os.Args[2:]...)
		cmd.Stderr = os.Stderr
		bytes, err := cmd.Output()

		if err != nil {
			if len(bytes) > 0 {
				util.Fatal(nil, "read command output failed: %s\nactual stdout: %s", err, string(bytes))
			}
			util.Fatal(nil, "read command output failed: %s", err)
		}

		if len(bytes) == 0 {
			util.Fatal(nil, "empty stdout, but without error")
		}

		if err := json.Unmarshal(bytes, &tmpCache); err != nil {
			util.Fatal(nil, "json.Unmarshal() failed(read command output): %s\nactual stdout: %s", err, string(bytes))
		}

		updated = true
		cache = tmpCache
	}

	// write phase (lock needed)
	if updated {
		// acquire lock
		lock := flock.New(cacheFilepath + ".lock")
		if err := lock.Lock(); err != nil {
			util.Fatal(nil, "file lock failed: %s", err)
		}

		// open file
		f, err := os.OpenFile(cacheFilepath, os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			// explicitly unlock since fatal will exit
			_ = lock.Unlock() // errcheck ignored on fatal path
			util.Fatal(nil, "file open failed: %s", err)
		}

		// re-read and merge latest data to avoid overwriting parallel changes
		freshBytes, err := io.ReadAll(f)
		if err != nil {
			f.Close()
			_ = lock.Unlock()
			util.Fatal(nil, "file read failed (during update): %s", err)
		}

		if len(freshBytes) > 0 {
			freshCache := CacheFile{}
			if err := json.Unmarshal(freshBytes, &freshCache); err == nil {
				// use the fresh cache map
				if len(freshCache.Credentials) > 0 {
					cacheFile.Credentials = freshCache.Credentials
				}
			}
		}

		// update cache key with the freshly acquired token
		if len(cacheFile.Credentials) == 0 {
			cacheFile.Credentials = map[string]ClientAuthentication{}
		}
		cacheFile.Credentials[cacheKey] = cache

		// cleanup expired tokens from the cache
		now := time.Now()
		for k, v := range cacheFile.Credentials {
			if now.After(v.Status.ExpirationTimestamp) {
				delete(cacheFile.Credentials, k)
			}
		}

		// write changes back to file
		if err := f.Truncate(0); err != nil {
			f.Close()
			_ = lock.Unlock()
			util.Fatal(nil, "file truncate failed: %s", err)
		}

		if _, err := f.Seek(0, 0); err != nil {
			f.Close()
			_ = lock.Unlock()
			util.Fatal(nil, "file seek failed: %s", err)
		}

		outBytes, err := json.Marshal(cacheFile)
		if err != nil {
			f.Close()
			_ = lock.Unlock()
			util.Fatal(nil, "json.Marshal() failed: %s", err)
		}

		if _, err := f.Write(outBytes); err != nil {
			f.Close()
			_ = lock.Unlock()
			util.Fatal(nil, "file write failed: %s", err)
		}

		f.Close()

		// release lock
		if err := lock.Unlock(); err != nil {
			util.Log("failed to unlock file: %s", err)
		}
	}

	// print
	output, err := json.Marshal(cache)
	if err != nil {
		util.Fatal(nil, "json.Marshal() failed: %s", err)
	}
	fmt.Println(string(output))
}
