package builder

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/BrightLocal/FrontBuilder/builder/files"
	"github.com/stretchr/testify/assert"
)

func TestPrepareApps(t *testing.T) {
	path, err := os.Getwd()
	if err != nil {
		log.Fatalf("error get working dir %s", err)
	}
	var testCases = map[string]Builder{
		"default_app": {
			htmlExtension: ".html",
			scripts: []sourcePath{
				{
					BaseDir: filepath.Join(path, "test-project", "default", "app"),
					Path:    "/app2.js",
				},
			},
			typeScripts: []sourcePath{
				{
					BaseDir: filepath.Join(path, "test-project", "default", "app"),
					Path:    "/index.ts",
				},
				{
					BaseDir: filepath.Join(path, "test-project", "default", "app"),
					Path:    "/app1.ts",
				},
			},
			htmls: map[string]*files.HTML{
				"/app1.html":  {},
				"/app2.html":  {},
				"/index.html": {},
			},
			jsApps: make(map[string]sourcePath),
		},
		"large_app": {
			htmlExtension: ".html",
			scripts: []sourcePath{
				{
					BaseDir: filepath.Join(path, "test-project", "large", "scripts"),
					Path:    "app2/app2.js",
				},
			},
			typeScripts: []sourcePath{
				{
					BaseDir: filepath.Join(path, "test-project", "large", "scripts"),
					Path:    "/index.ts",
				},
				{
					BaseDir: filepath.Join(path, "test-project", "large", "scripts"),
					Path:    "app1/app1.ts",
				},
			},
			htmls: map[string]*files.HTML{
				"app1/app1.html": {},
				"app2/app2.html": {},
				"/index.html":    {},
			},
			jsApps: make(map[string]sourcePath),
		},
		"minimal_app": {
			htmlExtension: ".htm",
			scripts:       []sourcePath{},
			typeScripts:   []sourcePath{},
			htmls: map[string]*files.HTML{
				"source/index.htm": {},
			},
			jsApps: make(map[string]sourcePath),
		},
	}
	var testResults = map[string]map[string]sourcePath{
		"default_app": {
			"/app1.html": sourcePath{
				BaseDir: filepath.Join(path, "test-project", "default", "app"),
				Path:    "/app1.ts",
			},
			"/app2.html": sourcePath{
				BaseDir: filepath.Join(path, "test-project", "default", "app"),
				Path:    "/app2.js",
			},
			"/index.html": sourcePath{
				BaseDir: filepath.Join(path, "test-project", "default", "app"),
				Path:    "/index.ts",
			},
		},
		"large_app": {
			"/index.html": sourcePath{
				BaseDir: filepath.Join(path, "test-project", "large", "scripts"),
				Path:    "/index.ts",
			},
			"app1/app1.html": sourcePath{
				BaseDir: filepath.Join(path, "test-project", "large", "scripts"),
				Path:    "app1/app1.ts",
			},
			"app2/app2.html": sourcePath{
				BaseDir: filepath.Join(path, "test-project", "large", "scripts"),
				Path:    "app2/app2.js",
			},
		},
		"minimal_app": {},
	}
	for i, tt := range testCases {
		t.Run(i, func(t *testing.T) {
			tt.prepareApps()
			assert.Equal(t, testResults[i], tt.jsApps)
		})
	}
}

func TestCollect(t *testing.T) {
	var testCases = map[string][]string{
		"default_app": {
			"../test-projects/default/app",
		},
		"large_app": {
			"../test-projects/large/scripts",
			"../test-projects/large/templates",
		},
		"minimal_app": {
			"../test-projects/minimal/source",
		},
	}
	var testResults = map[string]struct {
		sources     []string
		scripts     []sourcePath
		typeScripts []sourcePath
		htmls       int
	}{
		"default_app": {
			sources: []string{
				"../test-projects/default/app",
			},
			scripts: []sourcePath{
				{
					BaseDir: "../test-projects/default/app",
					Path:    "/app2.js",
				},
			},
			typeScripts: []sourcePath{
				{
					BaseDir: "../test-projects/default/app",
					Path:    "/app1.ts",
				},
				{
					BaseDir: "../test-projects/default/app",
					Path:    "/index.ts",
				},
			},
			htmls: 3,
		},
		"large_app": {
			sources: []string{
				"../test-projects/large/scripts",
				"../test-projects/large/templates",
			},
			scripts: []sourcePath{
				{
					BaseDir: "../test-projects/large/scripts",
					Path:    "/app2/app2.js",
				},
			},
			typeScripts: []sourcePath{
				{
					BaseDir: "../test-projects/large/scripts",
					Path:    "/app1/app1.ts",
				},
				{
					BaseDir: "../test-projects/large/scripts",
					Path:    "/index.ts",
				},
			},
			htmls: 3,
		},
		"minimal_app": {
			sources: []string{
				"../test-projects/minimal/source",
			},
			scripts:     nil,
			typeScripts: nil,
			htmls:       1,
		},
	}
	for name, sources := range testCases {
		t.Run(name, func(t *testing.T) {
			b := NewBuilder(sources, "", false)
			if name == "minimal_app" {
				b.HTMLExtension("htm")
			}
			if assert.NoError(t, b.collectFiles()) {
				assert.Equal(t, testResults[name].sources, b.sources)
				assert.Equal(t, testResults[name].scripts, b.scripts)
				assert.Equal(t, testResults[name].typeScripts, b.typeScripts)
				assert.Equal(t, testResults[name].htmls, len(b.htmls))
			}
		})
	}
}
