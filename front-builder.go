package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/BrightLocal/FrontBuilder/builder"
	"github.com/BrightLocal/FrontBuilder/watcher"
)

func main() {
	env := flag.String("env", "production", "specify what mode of build you need to use development or production")
	rootFolder := flag.String("folder", "./views", "specify root directory of frontend files")
	flag.Parse()
	workDir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	frontBuilder := builder.NewBuilder(*rootFolder, *env, workDir)
	frontBuilder.Build()
	if *env == "development" {
		buildWatcher, err := watcher.NewBuildWatcher(frontBuilder)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		buildWatcher.Watch()
	}
}
