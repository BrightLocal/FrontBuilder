package builder

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type HTMLFile struct {
	path   string
	script *JSFile
}

func NewHTMLFile(sourceFile string) *HTMLFile {
	return &HTMLFile{path: sourceFile}
}

func (h *HTMLFile) InjectJS(script *JSFile) *HTMLFile {
	h.script = script
	return h
}

func (h *HTMLFile) Render(destinationFile string, releaseBuild bool) error {
	htmlSrc, err := ioutil.ReadFile(h.path)
	if err != nil && err != io.EOF {
		return err
	}
	if h.script != nil {
		src := h.script.GetScriptSource(releaseBuild)
		if bytes.Contains(htmlSrc, []byte(`<script src=""></script>`)) {
			htmlSrc = bytes.Replace(htmlSrc, []byte(`src=""`), []byte(`src="/`+src+`"`), -1)
		}
	}
	if err := os.MkdirAll(filepath.Dir(destinationFile), 0770); err != nil {
		return err
	}
	if err := ioutil.WriteFile(destinationFile, htmlSrc, 0644); err != nil {
		return err
	}
	return nil
}
