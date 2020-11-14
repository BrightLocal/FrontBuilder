package builder

import (
	"testing"

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
		htmls: map[string]*HTMLFile{
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
