package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"runtime"
	"runtime/debug"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

var (
	inPlaceFlag bool
	restoreFlag bool
	replaceCmd  string = "kcc-cache"
)

func main() {
	// flag
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [flags] <kubeconfig-filepath>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.BoolVar(&inPlaceFlag, "i", false, "edit file in-place")
	flag.BoolVar(&restoreFlag, "r", false, "restore kubeconfig to original")
	flag.StringVar(&replaceCmd, "c", replaceCmd, "injection command")
	flag.Parse()

	// read input
	var filename string
	var bytes []byte
	{
		if args := flag.Args(); len(args) > 0 {
			filename = args[0]
		}
		if filename == "" {
			fmt.Fprintln(os.Stderr, "filename required")
			flag.Usage()
			os.Exit(1)
		}
		b, err := os.ReadFile(filename)
		if err != nil {
			fatal("read error: %s", err)
		}
		bytes = b
	}

	// parse as kubeconfig
	var kubeConfig api.Config
	{
		clientConfig, err := clientcmd.NewClientConfigFromBytes(bytes)
		if err != nil {
			fatal("%v", err)
		}

		apiConfig, err := clientConfig.RawConfig()
		if err != nil {
			fatal("%v", err)
		}
		kubeConfig = apiConfig
	}

	// kubeconfig manipulation
	if restoreFlag {
		// restore to original
		for _, user := range kubeConfig.AuthInfos {
			if user.Exec == nil {
				continue
			}
			if user.Exec.Command == replaceCmd {
				user.Exec.Command = user.Exec.Args[0]
				user.Exec.Args = user.Exec.Args[1:]
			}

			search := func() (index int) {
				for i, e := range user.Exec.Env {
					if e.Name == "KUBE_CREDENTIAL_CACHE_USER" {
						return i
					}
				}
				return -1
			}

			for {
				i := search()
				if i == -1 {
					break
				}
				user.Exec.Env = append(user.Exec.Env[:i], user.Exec.Env[i+1:]...)
			}
		}
	} else {
		// enable cache
		for name, user := range kubeConfig.AuthInfos {
			if user.Exec == nil {
				continue
			}
			if user.Exec.Command != replaceCmd {
				user.Exec.Args = append([]string{user.Exec.Command}, user.Exec.Args...)
				user.Exec.Command = replaceCmd
			}

			found := false
			userEnv := api.ExecEnvVar{
				Name:  "KUBE_CREDENTIAL_CACHE_USER",
				Value: name,
			}
			for i, e := range user.Exec.Env {
				if e.Name == "KUBE_CREDENTIAL_CACHE_USER" {
					found = true
					user.Exec.Env[i] = userEnv
				}
			}
			if !found {
				user.Exec.Env = append(user.Exec.Env, userEnv)
			}
		}
	}

	// output
	if inPlaceFlag {
		// in-place
		err := clientcmd.WriteToFile(kubeConfig, filename)
		if err != nil {
			fatal("%v", err)
		}
	} else {
		// stdout
		b, err := clientcmd.Write(kubeConfig)
		if err != nil {
			fatal("%v", err)
		}

		fmt.Println(string(b))
	}
}

func fatal(format string, v ...any) {
	var commit string = "main"
	if i, ok := debug.ReadBuildInfo(); ok {
		for _, v := range i.Settings {
			if v.Key == "vcs.revision" {
				commit = v.Value
			}
		}
	}
	_, _, line, _ := runtime.Caller(1)

	fmt.Fprintf(os.Stderr, "%s: ", path.Base(os.Args[0]))
	fmt.Fprintf(os.Stderr, format+"\n", v...)
	fmt.Fprintf(os.Stderr, "error occurred at: https://github.com/ryodocx/kube-credential-cache/blob/%s/cmd/kcc-injector/main.go#L%d\n", commit, line)
	flag.Usage()
	os.Exit(1)
}
