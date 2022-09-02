package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"runtime"
	"runtime/debug"
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
		cacheKeyEnvlist []string      = []string{"KUBE_CREDENTIAL_CACHE_USER", "AWS_PROFILE", "AWS_REGION", "AWS_VAULT"}
	)
	if e := os.Getenv("KUBE_CREDENTIAL_CACHE_FILE"); e != "" {
		cacheFilepath = e
	} else {
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			fatal("can't find CacheDir. fix error or set 'KUBE_CREDENTIAL_CACHE_FILE': %s", err)
		}
		cacheFilepath = path.Join(cacheDir, "kube-credential-cache", "cache.json")
	}
	if e := os.Getenv("KUBE_CREDENTIAL_CACHE_REFRESH_MARGIN"); e != "" {
		d, err := time.ParseDuration(e)
		if err != nil {
			fatal("invalid environment variable 'KUBE_CREDENTIAL_CACHE_REFRESH_MARGIN': %s", err.Error())
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
				fatal("mkdir failed: %s", err)
			}
			f, err = os.OpenFile(cacheFilepath, os.O_RDWR|os.O_CREATE, 0600)
			if err != nil {
				fatal("file open failed(after mkdir): %s", err)
			}
		} else {
			fatal("file open failed: %s", err)
		}
	}
	defer f.Close()

	// read file
	updated := false
	cacheFile := CacheFile{}
	bytes, err := io.ReadAll(f)
	if err != nil {
		fatal("file read failed: %s", err)
	}
	if len(bytes) > 0 {
		if err := json.Unmarshal(bytes, &cacheFile); err != nil {
			log("json.Unmarshal() failed(read cache file): %s\n...Corruption detected, recreate cache file", err)
			updated = true
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

			// cleanup
			for k, v := range cacheFile.Credentials {
				if time.Now().After(v.Status.ExpirationTimestamp) {
					delete(cacheFile.Credentials, k)
				}
			}

			// update
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
			fatal("not enough command at args")
		}
		cmd := exec.Command(os.Args[1], os.Args[2:]...)
		cmd.Stderr = os.Stderr
		bytes, err := cmd.Output()

		if err != nil {
			if len(bytes) > 0 {
				fatal("read command output failed: %s\nactual stdout: %s", err, string(bytes))
			}
			fatal("read command output failed: %s", err)
		}

		if len(bytes) == 0 {
			fatal("empty stdout, but without error")
		}

		if err := json.Unmarshal(bytes, &tmpCache); err != nil {
			fatal("json.Unmarshal() failed(read command output): %s\nactual stdout: %s", err, string(bytes))
		}

		cacheFile.Credentials[cacheKey] = tmpCache
		updated = true
	}

	// print
	output, err := json.Marshal(cacheFile.Credentials[cacheKey])
	if err != nil {
		fatal("json.Marshal() failed: %s", err)
	}
	fmt.Println(string(output))
}

func fatal(format string, v ...any) {
	log(format, v...)

	var commit string = "main"
	if i, ok := debug.ReadBuildInfo(); ok {
		for _, v := range i.Settings {
			if v.Key == "vcs.revision" {
				commit = v.Value
			}
		}
	}
	_, _, line, _ := runtime.Caller(1)
	fmt.Fprintf(os.Stderr, "error occurred at: https://github.com/ryodocx/kube-credential-cache/blob/%s/cmd/kcc-cache/main.go#L%d\n", commit, line)

	os.Exit(1)
}

func log(format string, v ...any) {
	fmt.Fprintf(os.Stderr, "%s: ", path.Base(os.Args[0]))
	fmt.Fprintf(os.Stderr, format+"\n", v...)
}
