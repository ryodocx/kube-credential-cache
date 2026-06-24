package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/ryodocx/kube-credential-cache/internal/util"
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
			util.Fatal(flag.Usage, "read error: %s", err)
		}
		bytes = b
	}

	// parse as kubeconfig
	var kubeConfig api.Config
	{
		clientConfig, err := clientcmd.NewClientConfigFromBytes(bytes)
		if err != nil {
			util.Fatal(flag.Usage, "%v", err)
		}

		apiConfig, err := clientConfig.RawConfig()
		if err != nil {
			util.Fatal(flag.Usage, "%v", err)
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

			var newEnv []api.ExecEnvVar
			for _, e := range user.Exec.Env {
				if e.Name != "KUBE_CREDENTIAL_CACHE_USER" {
					newEnv = append(newEnv, e)
				}
			}
			user.Exec.Env = newEnv
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
			util.Fatal(flag.Usage, "%v", err)
		}
	} else {
		// stdout
		b, err := clientcmd.Write(kubeConfig)
		if err != nil {
			util.Fatal(flag.Usage, "%v", err)
		}

		fmt.Println(string(b))
	}
}
