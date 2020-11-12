package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/BrightLocal/FrontBuilder/builder"
	"github.com/BrightLocal/FrontBuilder/watcher"
)

func main() {
	cfg := configure()
	frontBuilder := builder.NewBuilder(cfg.Source, cfg.Env, cfg.Destination)
	frontBuilder.Build()
	if cfg.Watch {
		buildWatcher, err := watcher.NewBuildWatcher(frontBuilder)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		buildWatcher.Watch()
	}
}

type config struct {
	Env         string
	Watch       bool
	Source      string
	Destination string
}

func (c config) IsProduction() bool {
	return strings.HasPrefix(c.Env, "prod")
}

func configure() config {
	cfg := config{
		Env:         "production",
		Watch:       false,
		Source:      "./views",  // TODO make configurable
		Destination: "./static", // TODO make configurable
	}
	if len(os.Args) == 2 {
		cfg.Watch = os.Args[1] == "watch"
	} else if len(os.Args) == 3 {
		if os.Args[1] != "build" {
			fmt.Println("Expected command: 'build' or 'watch'")
			usage()
			os.Exit(1)
		}
		if e := os.Args[2]; e != "" {
			cfg.Env = e
		}
	} else {
		usage()
	}
	return cfg
}

func usage() {
	fmt.Printf(`Usage:
%[1]s build prod -- builds production version
%[1]s build      -- same as 'build prod'
%[1]s build dev  -- builds development version
%[1]s watch      -- build dev version and continue watching for files change
`, path.Base(os.Args[0]))
	os.Exit(0)
}
