package builder

type JSFile struct {
	path    string
	content []byte
}

func NewJSFile(path string, content []byte) *JSFile {
	return &JSFile{path: path, content: content}
}

func (j *JSFile) GetScriptSource(releaseBuild bool) string {
	// TODO Implement me
	return ""
}
