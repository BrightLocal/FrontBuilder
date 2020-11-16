package builder

import (
	"testing"

	"github.com/BrightLocal/FrontBuilder/builder/files"
	"github.com/stretchr/testify/assert"
)

func TestPrepareApps(t *testing.T) {
	b := Builder{
		htmlExtension: ".html",
		scripts: []string{
			"app/index.js",
			"app/lib.js",
		},
		typeScripts: []string{
			"index.ts",
		},
		htmls: map[string]*files.HTML{
			"index.html":     {},
			"app/index.html": {},
		},
		jsApps: make(map[string]string),
	}
	b.prepareApps()
	assert.Equal(t, map[string]string{
		"index.html":     "index.ts",
		"app/index.html": "app/index.js",
	}, b.jsApps)
}

func TestCollect(t *testing.T) {
	sources := []string{
		"../test-projects/large/scripts",
		"../test-projects/large/templates",
	}
	b := NewBuilder(sources, "", false)
	if assert.NoError(t, b.collectFiles()) {
		assert.Equal(t, []string{
			"../test-projects/large/scripts/",
			"../test-projects/large/templates/",
		}, b.sources)
		assert.Equal(t, []string{"app2/app2.js"}, b.scripts)
		assert.Equal(t, []string{"app1/app1.ts", "index.ts"}, b.typeScripts)
		assert.Equal(t, 3, len(b.htmls))
	}
}
