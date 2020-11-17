package files

import (
	"crypto/md5"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type JS struct {
	destination string
	builtScript string
	content     []byte
}

func NewJS(destination, scriptFile string, content []byte) *JS {
	return &JS{
		destination: destination,
		builtScript: scriptFile,
		content:     content,
	}
}

func (j *JS) GetScriptSource(releaseBuild bool) (string, error) {
	filePath := "/" + strings.TrimPrefix(j.builtScript, j.destination)
	if !releaseBuild {
		return filePath, nil
	}
	ext := path.Ext(filePath)
	hash := md5.Sum(j.content)
	source := fmt.Sprintf("%s.%x%s",
		strings.TrimSuffix(filePath, ext),
		hash[:4],
		ext,
	)
	if err := os.Rename(j.builtScript, filepath.Join(j.destination, source)); err != nil {
		return "", err
	}
	return source, nil
}
