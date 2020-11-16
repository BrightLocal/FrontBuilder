package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Config struct {
	Env           string
	Watch         bool
	Source        []string
	Destination   string
	IndexFile     string
	HTMLExtension string
	ScriptsPrefix string
	HTMLPrefix    string
}

func Configure() Config {
	cfg := Config{
		Env:   "production",
		Watch: false,
	}
	if len(os.Args) == 2 {
		if os.Args[1] == "watch" {
			cfg.Env = "development"
			cfg.Watch = true
		}
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
	if err := cfg.readConfigFile(); err != nil {
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

func (c Config) IsProduction() bool {
	return strings.HasPrefix(c.Env, "prod")
}

func (c *Config) readConfigFile() error {
	f, err := os.Open(".front-builder.json")
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	type fConfig struct {
		Source        interface{} `json:"source"`
		Destination   string      `json:"destination"`
		IndexFile     string      `json:"index_file"`
		HTMLExtension string      `json:"html_extension"`
		ScriptsPrefix string      `json:"scripts_prefix"`
		HTMLPrefix    string      `json:"html_prefix"`
	}
	var fc fConfig
	if err = json.NewDecoder(f).Decode(&fc); err != nil {
		return err
	}
	if src, ok := fc.Source.(string); ok {
		c.Source = []string{src}
	} else if src, ok := fc.Source.([]interface{}); ok {
		for _, s := range src {
			if s, ok := s.(string); ok {
				c.Source = append(c.Source, s)
			} else {
				return errors.New("source can be either string or array of strings")
			}
		}
	} else {
		return errors.New("source can be either string or array of strings")
	}
	c.Destination = fc.Destination
	c.IndexFile = fc.IndexFile
	c.HTMLExtension = fc.HTMLExtension
	c.ScriptsPrefix = fc.ScriptsPrefix
	c.HTMLPrefix = fc.HTMLPrefix
	return nil
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
