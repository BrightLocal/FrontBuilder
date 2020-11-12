package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/BrightLocal/FrontBuilder/builder"
	"github.com/BrightLocal/FrontBuilder/watcher"
)

func main() {
	cfg := configure()
	frontBuilder := builder.NewBuilder(cfg.Source[0], cfg.Env, cfg.Destination)
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
	Source      []string
	Destination string
}

func (c config) IsProduction() bool {
	return strings.HasPrefix(c.Env, "prod")
}

func (c *config) readFile() error {
	f, err := os.Open(".front-builder.json")
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	type fConfig struct {
		Source      interface{}
		Destination string
	}
	var fc fConfig
	if err = json.NewDecoder(f).Decode(&fc); err != nil {
		return err
	}
	if src, ok := fc.Source.(string); ok {
		c.Source = []string{src}
	} else if src, ok := fc.Source.([]string); ok {
		c.Source = src
	} else {
		return errors.New("source can be either string or array of strings")
	}
	c.Destination = fc.Destination
	return nil
}

func configure() config {
	cfg := config{
		Env:   "production",
		Watch: false,
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
	if err := cfg.readFile(); err != nil {
		fmt.Printf("Error reading config file: %s\n", err)
		os.Exit(1)
	}
	var err error
	if cfg.Destination, err = filepath.Abs(cfg.Destination); err != nil {
		fmt.Printf("Error expanind destination path: %s\n", err)
		os.Exit(1)
	}
	if _, err = os.Stat(cfg.Destination); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(cfg.Destination, 0750); err != nil {
			fmt.Printf("Error creating missing destination directory: %s", err)
			os.Exit(1)
		}
	}
	for i := range cfg.Source {
		if cfg.Source[i], err = filepath.Abs(cfg.Source[i]); err != nil {
			fmt.Printf("Error expanind source path %q: %s\n", cfg.Source[i], err)
			os.Exit(1)
		}
		if _, err := os.Stat(cfg.Source[i]); err != nil && os.IsNotExist(err) {
			fmt.Printf("Source directory %q does not exists: %s", cfg.Source[i], err)
			os.Exit(1)
		}
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
