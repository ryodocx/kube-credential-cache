package benchmark

import (
	"os/exec"
	"testing"
)

func BenchmarkKubectlDirect(b *testing.B) {
	run := func() {
		_, err := exec.Command(
			"kubectl",
			"version",
			"--user",
			"kind-kcc-bench",
		).Output()
		if err != nil {
			b.Error(err)
		}
	}

	run()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		run()
	}
}

func BenchmarkKubectlCache(b *testing.B) {
	run := func() {
		_, err := exec.Command(
			"kubectl",
			"version",
			"--user",
			"cache",
		).Output()
		if err != nil {
			b.Error(err)
		}
	}

	run()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		run()
	}
}

func BenchmarkKubectlSlow(b *testing.B) {
	run := func() {
		_, err := exec.Command(
			"kubectl",
			"version",
			"--user",
			"slow",
		).Output()
		if err != nil {
			b.Error(err)
		}
	}

	run()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		run()
	}
}

func BenchmarkGetTokenCache(b *testing.B) {
	run := func() {
		_, err := exec.Command(
			"kcc-cache",
			"sh",
			"get-token.sh",
		).Output()
		if err != nil {
			b.Error(err)
		}
	}

	run()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		run()
	}
}

func BenchmarkGetToken(b *testing.B) {
	run := func() {
		_, err := exec.Command(
			"sh",
			"get-token.sh",
		).Output()
		if err != nil {
			b.Error(err)
		}
	}

	run()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		run()
	}
}
