package files

import (
	"crypto/md5"
	"fmt"
	"path"
	"strings"
)

type JS struct {
	dst     string
	content []byte
}

func NewJS(destinationFile string, content []byte) *JS {
	return &JS{
		dst:     "/" + strings.TrimLeft(destinationFile, "/"),
		content: content,
	}
}

func (j *JS) GetScriptSource(releaseBuild bool) string {
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
