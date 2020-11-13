package builder

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTML(t *testing.T) {
	html := NewHTMLFile("test_files/test.html").
		InjectJS(NewJSFile("script.js", []byte("console.log('hello');\n")))
	if assert.NoError(t, html.Render("test_files/out.html", false)) {
		if r, err := ioutil.ReadFile("test_files/out.html"); assert.NoError(t, err) {
			expectHTML := []byte("Boo ya!\n<script src=\"/script.js\"></script>\nBye!\n")
			if assert.Equal(t, r, expectHTML) {
				_ = os.Remove("test_files/out.html")
			}
		}
	}
	if assert.NoError(t, html.Render("test_files/out.html", true)) {
		if r, err := ioutil.ReadFile("test_files/out.html"); assert.NoError(t, err) {
			expectHTML := []byte("Boo ya!\n<script src=\"/script.HASH-HERE.js\"></script>\nBye!\n")
			if assert.Equal(t, r, expectHTML) {
				_ = os.Remove("test_files/out.html")
			}
		}
	}
}
