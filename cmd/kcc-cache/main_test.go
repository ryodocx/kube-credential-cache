package main

import (
	"bytes"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
)

func TestMainMissingArgsError(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		os.Args = []string{"kcc-cache"}
		main()
		return
	}

	// Create a temporary directory for the cache file to avoid polluting user environment
	tmpDir := t.TempDir()
	cacheFilePath := path.Join(tmpDir, "test_cache.json")

	cmd := exec.Command(os.Args[0], "-test.run=TestMainMissingArgsError")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1", "KUBE_CREDENTIAL_CACHE_FILE="+cacheFilePath)

	// We want to capture stderr to check the error message
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()

	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		output := stderr.String()
		if !strings.Contains(output, "not enough command at args") {
			t.Errorf("expected error message to contain 'not enough command at args', got: %s", output)
		}
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}
