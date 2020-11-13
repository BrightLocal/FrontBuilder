package builder

type JSFile struct {
	path    string
	content []byte
}

func NewJSFile(destinationFile string, content []byte) *JSFile {
	return &JSFile{path: destinationFile, content: content}
}

func (j *JSFile) GetScriptSource(releaseBuild bool) string {
	// TODO Implement me
	return ""
}
