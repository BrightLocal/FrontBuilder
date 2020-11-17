package files

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTML(t *testing.T) {
	const root = "./test_files"
	testCases := []struct {
		htmlFile     string
		scriptFile   string
		outFile      string
		release      bool
		expectToFind string
	}{
		{
			htmlFile:     "/source.html",
			scriptFile:   "/script.js",
			outFile:      "out.html",
			release:      false,
			expectToFind: `src="/script.js"`,
		},
		{
			htmlFile:     "/source.html",
			scriptFile:   "/script.js",
			outFile:      "out.html",
			release:      true,
			expectToFind: `src="/script.cd4d3d46.js"`,
		},
	}
	for _, tt := range testCases {
		if script, err := ioutil.ReadFile(root + tt.scriptFile); assert.NoError(t, err) {
			html := NewHTML(root + tt.htmlFile)
			html.InjectJS(NewJS(root, root+tt.scriptFile, script))
			if assert.NoError(t, html.Render(root+tt.outFile, tt.release)) {
				if r, err := ioutil.ReadFile(root + tt.outFile); assert.NoError(t, err) {
					assert.True(t, bytes.Contains(r, []byte(tt.expectToFind)), string(r))
					_ = os.Remove(root + tt.outFile)
				}
			}
		}
	}
	if err := os.Rename(root+"/script.cd4d3d46.js", root+"/script.js"); err != nil {
		log.Fatal(err)
	}
}
