package builder

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
)

type HTMLFile struct {
	src    string
	script *JSFile
}

var appPlaceholder = []byte(`<!--#APP#-->`)

func NewHTMLFile(sourceFile string) *HTMLFile {
	return &HTMLFile{src: sourceFile}
}

func (h *HTMLFile) InjectJS(script *JSFile) *HTMLFile {
	h.script = script
	return h
}

func (h *HTMLFile) Render(destinationFile string, releaseBuild bool) error {
	html, err := ioutil.ReadFile(h.src)
	if err != nil {
		return err
	}
	if s := h.script; s != nil {
		html = bytes.ReplaceAll(
			html,
			appPlaceholder,
			[]byte(`<script src="`+s.GetScriptSource(releaseBuild)+`"></script>`),
		)
	}
	if err := os.MkdirAll(filepath.Dir(destinationFile), 0750); err != nil {
		return err
	}
	return ioutil.WriteFile(destinationFile, html, 0640)
}
