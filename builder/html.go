package builder

type HTMLFile struct {
	path   string
	jsFile *JSFile
}

func NewHTMLFile(path string) *HTMLFile {
	return &HTMLFile{path: path}
}

func (h *HTMLFile) Render(destination string, releaseBuild bool) error {
	// TODO implement me
	if h.jsFile != nil {
		h.jsFile.GetScriptSource(releaseBuild)
	}
	return nil
}

func (h *HTMLFile) InjectJS(jsFile *JSFile) {
	h.jsFile = jsFile
}
