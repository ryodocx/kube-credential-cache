package main

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
)

// TestParseInvalidKubeConfig tests the exact lines in main.go
// that load bytes into a clientcmd config and parse the RawConfig.
// We test this using the standard go sub-process execution pattern
// to verify fatal() calls os.Exit(1).
func TestParseInvalidKubeConfig(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		// Replace args with our test file
		os.Args = []string{"kcc-injector", "invalid-kubeconfig.yaml"}
		main()
		return
	}

	// Create an invalid yaml/json file
	err := os.WriteFile("invalid-kubeconfig.yaml", []byte("invalid content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove("invalid-kubeconfig.yaml")

	cmd := exec.Command(os.Args[0], "-test.run=TestParseInvalidKubeConfig")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		// verify the error message relates to the parsing failure
		out := stderr.String()
		if !bytes.Contains([]byte(out), []byte("couldn't get version/kind")) {
			t.Errorf("expected parsing error in output, got: %s", out)
		}
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}
