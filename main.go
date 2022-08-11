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

func main() {
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
		bytes, err := cmd.CombinedOutput()

		if err != nil {
			log.Fatalf("read command output failed: %s\noutput: %s", err, string(bytes))
		}

		if err := json.Unmarshal(bytes, &tmpCache); err != nil {
			log.Fatalf("json.Unmarshal() failed(read command output): %s", err)
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
