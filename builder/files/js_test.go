package files

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSFile(t *testing.T) {
	const root = "./test_files/"
	j := NewJS(root, root+"script.js", []byte("console.log('hello');\n"))
	script, err := j.GetScriptSource(true)
	if err != nil {
		assert.Fail(t, err.Error())
	}
	assert.Equal(t, "/script.cd4d3d46.js", script)
	script, err = j.GetScriptSource(false)
	if err != nil {
		assert.Fail(t, err.Error())
	}
	assert.Equal(t, "/script.js", script)
	if err := os.Rename(root+"/script.cd4d3d46.js", root+"/script.js"); err != nil {
		log.Fatal(err)
	}
}
