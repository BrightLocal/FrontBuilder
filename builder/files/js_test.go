package files

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSFile(t *testing.T) {
	j := NewJS("path/to/script.js", []byte("console.log('hello');\n"))
	assert.Equal(t, "/path/to/script.cd4d3d46.js", j.GetScriptSource(true))
	assert.Equal(t, "/path/to/script.js", j.GetScriptSource(false))
}
