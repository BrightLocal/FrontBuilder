package builder

import (
	"crypto/md5"
	"fmt"
	"path"
	"path/filepath"
)

type JSFile struct {
	path    string
	content []byte
}

func NewJSFile(destinationFile string, content []byte) *JSFile {
	return &JSFile{path: destinationFile, content: content}
}

func (j *JSFile) GetScriptSource(releaseBuild bool) string {
	_, fileName := filepath.Split(j.path)
	if releaseBuild {
		hashSum := md5.Sum(j.content)
		fileHash := fmt.Sprintf("%x", hashSum)[:8]
		ext := path.Ext(fileName)
		outfile := fileName[0:len(fileName)-len(ext)] + "." + fileHash + ".js"
		return outfile
	}
	return fileName
}
