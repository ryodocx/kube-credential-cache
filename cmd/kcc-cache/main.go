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
		cacheKeyEnvlist []string      = []string{"AWS_PROFILE", "AWS_REGION"}
	)
	if e := os.Getenv("KUBE_CREDENTIAL_CACHE_FILE"); e != "" {
		cacheFilepath = e
	} else {
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			log.Fatalf(os.Args[0]+": can't find CacheDir. fix error or set 'KUBE_CREDENTIAL_CACHE_FILE': %s", err)
		}
		cacheFilepath = path.Join(cacheDir, "kube-credential-cache", "cache.json")
	}
	if e := os.Getenv("KUBE_CREDENTIAL_CACHE_REFRESH_MARGIN"); e != "" {
		d, err := time.ParseDuration(e)
		if err != nil {
			log.Fatalf(os.Args[0]+": invalid environment variable 'KUBE_CREDENTIAL_CACHE_REFRESH_MARGIN': %s", err.Error())
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

	// open file
	f, err := os.OpenFile(cacheFilepath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(path.Dir(cacheFilepath), 0700); err != nil {
				log.Fatalf(os.Args[0]+": mkdir failed: %s", err)
			}
			f, err = os.OpenFile(cacheFilepath, os.O_RDWR|os.O_CREATE, 0600)
			if err != nil {
				log.Fatalf(os.Args[0]+": file open failed(after mkdir): %s", err)
			}
		} else {
			log.Fatalf(os.Args[0]+": file open failed: %s", err)
		}
	}
	defer f.Close()

	// read file
	updated := false
	cacheFile := CacheFile{}
	bytes, err := io.ReadAll(f)
	if err != nil {
		log.Fatalf(os.Args[0]+": file read failed: %s", err)
	}
	if len(bytes) > 0 {
		if err := json.Unmarshal(bytes, &cacheFile); err != nil {
			log.Fatalf(os.Args[0]+": json.Unmarshal() failed(read cache file): %s", err)
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
	if len(cacheFile.Credentials) == 0 {
		cacheFile.Credentials = map[string]ClientAuthentication{}
	}
	cache, ok := cacheFile.Credentials[cacheKey]
	if !ok || ok && time.Until(cache.Status.ExpirationTimestamp) < refreshMargin {
		// refresh
		tmpCache := ClientAuthentication{}

		if len(os.Args) < 2 {
			log.Fatalf(os.Args[0] + ": not enough command at args")
		}
		cmd := exec.Command(os.Args[1], os.Args[2:]...)
		cmd.Stderr = os.Stderr
		bytes, err := cmd.Output()

		if err != nil {
			log.Fatalf(os.Args[0]+": read command output failed: %s\noutput: %s", err, string(bytes))
		}

		if err := json.Unmarshal(bytes, &tmpCache); err != nil {
			log.Fatalf(os.Args[0]+": json.Unmarshal() failed(read command output): %s\nactual stdout: %s", err, string(bytes))
		}

		if time.Until(tmpCache.Status.ExpirationTimestamp) < refreshMargin {
			log.Fatalf(os.Args[0]+": Obtained token has expired: %s", string(bytes))
		}

		cacheFile.Credentials[cacheKey] = tmpCache
		updated = true
	}

	// print
	output, err := json.Marshal(cacheFile.Credentials[cacheKey])
	if err != nil {
		log.Fatalf(os.Args[0]+": json.Marshal() failed: %s", err)
	}
	fmt.Println(string(output))
}
