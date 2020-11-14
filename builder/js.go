package builder

import (
	"crypto/md5"
	"fmt"
	"path"
	"strings"
)

type JSFile struct {
	dst     string
	content []byte
}

func NewJSFile(destinationFile string, content []byte) *JSFile {
	return &JSFile{
		dst:     "/" + strings.TrimLeft(destinationFile, "/"),
		content: content,
	}
}

func (j *JSFile) GetScriptSource(releaseBuild bool) string {
	if !releaseBuild {
		return j.dst
	}
	ext := path.Ext(j.dst)
	hash := md5.Sum(j.content)
	return fmt.Sprintf("%s.%x%s",
		strings.TrimSuffix(j.dst, ext),
		hash[:4],
		ext,
	)
}
