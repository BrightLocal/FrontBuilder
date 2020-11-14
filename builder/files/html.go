package files

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
)

type HTML struct {
	src    string
	script *JS
}

var appPlaceholder = []byte(`<!--#APP#-->`)

func NewHTML(sourceFile string) *HTML {
	return &HTML{src: sourceFile}
}

func (h *HTML) InjectJS(script *JS) *HTML {
	h.script = script
	return h
}

func (h *HTML) Render(destinationFile string, releaseBuild bool) error {
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
