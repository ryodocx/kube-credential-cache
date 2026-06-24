package util

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"runtime/debug"
	"strings"
)

// Fatal prints the error message, the error location, optionally calls usage, and exits.
func Fatal(usage func(), format string, v ...any) {
	fmt.Fprintf(os.Stderr, "%s: ", path.Base(os.Args[0]))
	fmt.Fprintf(os.Stderr, format+"\n", v...)

	var commit string = "main"
	if i, ok := debug.ReadBuildInfo(); ok {
		for _, v := range i.Settings {
			if v.Key == "vcs.revision" {
				commit = v.Value
			}
		}
	}
	_, file, line, _ := runtime.Caller(1)

	repoPath := file
	if idx := strings.Index(file, "kube-credential-cache/"); idx != -1 {
		repoPath = file[idx+len("kube-credential-cache/"):]
	} else if idx := strings.LastIndex(file, "cmd/"); idx != -1 {
		repoPath = file[idx:]
	} else if idx := strings.LastIndex(file, "app/"); idx != -1 {
		repoPath = file[idx+len("app/"):]
	}

	fmt.Fprintf(os.Stderr, "error occurred at: https://github.com/ryodocx/kube-credential-cache/blob/%s/%s#L%d\n", commit, repoPath, line)

	if usage != nil {
		usage()
	}

	os.Exit(1)
}

// Log prints a message with the executable name as a prefix.
func Log(format string, v ...any) {
	fmt.Fprintf(os.Stderr, "%s: ", path.Base(os.Args[0]))
	fmt.Fprintf(os.Stderr, format+"\n", v...)
}
