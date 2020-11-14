package builder

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTML(t *testing.T) {
	const root = "test_files/"
	testCases := []struct {
		htmlFile     string
		scriptFile   string
		outFile      string
		release      bool
		expectToFind string
	}{
		{
			htmlFile:     "source.html",
			scriptFile:   "script.js",
			outFile:      "out.html",
			release:      false,
			expectToFind: `src="/test_files/script.js"`,
		},
		{
			htmlFile:     "source.html",
			scriptFile:   "script.js",
			outFile:      "out.html",
			release:      true,
			expectToFind: `src="/test_files/script.cd4d3d46.js"`,
		},
	}
	for _, tt := range testCases {
		if script, err := ioutil.ReadFile(root + tt.scriptFile); assert.NoError(t, err) {
			html := NewHTMLFile(root + tt.htmlFile)
			html.InjectJS(
				NewJSFile(root+tt.scriptFile, script),
			)
			if assert.NoError(t, html.Render(root+tt.outFile, tt.release)) {
				if r, err := ioutil.ReadFile(root + tt.outFile); assert.NoError(t, err) {
					assert.True(t, bytes.Contains(r, []byte(tt.expectToFind)), string(r))
					_ = os.Remove(root + tt.outFile)
				}
			}
		}
	}
}
