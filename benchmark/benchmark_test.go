package benchmark

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
)

// EKS
var eksKubectx string = os.Getenv("KUBECONTEXT_EKS")
var eksKubectxCache string = os.Getenv("KUBECONTEXT_EKS_CACHE")
var eksEnv map[string]string = map[string]string{
	"AWS_PROFILE": os.Getenv("AWS_PROFILE_EKS"),
	"KUBECONFIG":  os.Getenv("KUBECONFIG_EKS"),
}

func BenchmarkKubectlEKS(b *testing.B) {
	common(b, []string{
		"kubectl",
		"--context",
		eksKubectx,
		"version",
	}, eksEnv)
}

func BenchmarkKubectlEKSCache(b *testing.B) {
	common(b, []string{
		"kubectl",
		"--context",
		eksKubectxCache,
		"version",
	}, eksEnv)
}

func BenchmarkGetCredentialEKS(b *testing.B) {
	common(b, []string{
		"aws",
		"eks",
		"get-token",
		"--cluster-name",
		"example", // eks get-token は署名付きURLを生成しているのみなので、通信をせず任意のcluste-nameで動作する
	}, eksEnv)
}

func BenchmarkGetCredentialEKSCache(b *testing.B) {
	common(b, []string{
		"kcc-cache",
		"aws",
		"eks",
		"get-token",
		"--cluster-name",
		"example", // eks get-token は署名付きURLを生成しているのみなので、通信をせず任意のcluste-nameで動作する
	}, eksEnv)
}

// no wait

func BenchmarkKubectlNoWait(b *testing.B) {
	common(b, []string{
		"kubectl",
		"version",
		"--user",
		"kind-kcc-bench",
	})
}

// get-credential-wait.sh with cache

func BenchmarkKubectlCache(b *testing.B) {
	common(b, []string{
		"kubectl",
		"version",
		"--user",
		"cache",
	})
}

func BenchmarkGetCredentialCache(b *testing.B) {
	common(b, []string{
		"kcc-cache",
		"sh",
		"get-credential-wait.sh",
	})
}

// get-credential-wait.sh only

func BenchmarkKubectlSlow(b *testing.B) {
	common(b, []string{
		"kubectl",
		"version",
		"--user",
		"slow",
	})
}

func BenchmarkGetCredentialSlow(b *testing.B) {
	common(b, []string{
		"sh",
		"get-credential-wait.sh",
	})
}

// common
func common(b *testing.B, cmds []string, env ...map[string]string) {
	run := func() {
		cmd := exec.Command(cmds[0], cmds[1:]...)
		cmd.Env = os.Environ()
		if len(env) == 1 {
			for k, v := range env[0] {
				cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
			}
		}
		_, err := cmd.Output()
		if err != nil {
			b.Error(err)
		}
	}

	run() // warmup
	time.Sleep(time.Second * 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		run()
	}
}
