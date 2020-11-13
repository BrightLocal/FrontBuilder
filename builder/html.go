package builder

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
	// TODO implement me
	if h.script != nil {
		src := h.script.GetScriptSource(releaseBuild)
		_ = src
	}
	return nil
}
