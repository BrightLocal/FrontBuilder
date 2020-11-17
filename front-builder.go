package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/BrightLocal/FrontBuilder/builder"
	"github.com/BrightLocal/FrontBuilder/config"
	"github.com/BrightLocal/FrontBuilder/watcher"
)

func main() {
	start := time.Now()
	log.Println("Start build process")
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
	if err := frontBuilder.Build(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	log.Printf("Build finished: %s", time.Since(start))
	if cfg.Watch {
		buildWatcher, err := watcher.NewBuildWatcher()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		allEvents := make(chan struct{})
		for _, src := range cfg.Source {
			events, err := buildWatcher.Watch(src)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			go func(e chan struct{}) {
				for event := range e {
					allEvents <- event
				}
			}(events)
		}
		for range allEvents {
			if err = frontBuilder.Build(); err != nil {
				log.Printf("error rebuilding files: %s", err)
			}
		}
	}
}
