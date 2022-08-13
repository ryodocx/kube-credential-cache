package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

var (
	inPlaceFlag bool
	replaceCmd  string = "kcc-cache"
)

func main() {
	// flag
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [flags] <kubeconfig filepath>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.BoolVar(&inPlaceFlag, "i", false, "edit file in-place")
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
			log.Fatalln("filename required")
		}
		b, err := os.ReadFile(filename)
		if err != nil {
			log.Fatalf("read error: %s\n", err)
		}
		bytes = b
	}

	// parse as kubeconfig
	var kubeConfig api.Config
	{
		clientConfig, err := clientcmd.NewClientConfigFromBytes(bytes)
		if err != nil {
			log.Fatalln(err)
		}

		apiConfig, err := clientConfig.RawConfig()
		if err != nil {
			log.Fatalln(err)
		}
		kubeConfig = apiConfig
	}

	// manipulation
	for _, user := range kubeConfig.AuthInfos {
		if user.Exec == nil || user.Exec.Command == replaceCmd {
			continue
		}
		user.Exec.Args = append([]string{user.Exec.Command}, user.Exec.Args...)
		user.Exec.Command = replaceCmd
	}

	// output
	if inPlaceFlag {
		// in-place
		err := clientcmd.WriteToFile(kubeConfig, filename)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		// stdout
		b, err := clientcmd.Write(kubeConfig)
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println(string(b))
	}
}
