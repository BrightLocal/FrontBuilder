package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/BrightLocal/FrontBuilder/builder"
	"github.com/BrightLocal/FrontBuilder/config"
	"github.com/BrightLocal/FrontBuilder/watcher"
)

func main() {
	start := time.Now()
	fmt.Println("Start build process")
	cfg := config.Configure()
	frontBuilder := builder.NewBuilder(cfg.Source, cfg.Destination, cfg.IsProduction())
	if cfg.HTMLExtension != "" {
		frontBuilder.HTMLExtension(cfg.HTMLExtension)
	}
	if cfg.IndexFile != "" {
		frontBuilder.IndexFile(cfg.IndexFile)
	}
	frontBuilder.ScriptsPrefix(cfg.ScriptsPrefix)
	frontBuilder.HTMLPrefix(cfg.HTMLPrefix)
	frontBuilder.TypeScriptConfig(cfg.TypeScriptConfig)
	if err := frontBuilder.Build(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("Build finished: %s\n", time.Since(start))
	if cfg.Watch {
		buildWatcher, err := watcher.NewBuildWatcher(cfg.Source)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		done := make(chan struct{})
		events, err := buildWatcher.Watch()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		go func(e chan string) {
			for cmd := range e {
				log.Println("Rebuild project files")
				strs := strings.Split(cmd, ":")
				log.Printf("str0: %s; str1: %s", strs[0], strs[1])
				if strs[1] == "WRITE" {
					if err = frontBuilder.Rebuild(); err != nil {
						log.Printf("error rebuilding files: %s", err)
					}
				} else if strings.HasSuffix(strs[0], "~") {
					if err = frontBuilder.Build(); err != nil {
						log.Printf("error rebuilding files: %s", err)
					}
				}
			}
		}(events)
		<-done
	}
}
