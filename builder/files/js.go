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
	dst     string
	content []byte
}

func NewJS(destinationFile string, content []byte) *JS {
	return &JS{
		dst:     "/" + destinationFile,
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

func (j *JS) Rename(destination string, releaseBuild bool) error {
	if !releaseBuild {
		return nil
	}
	if err := os.Rename(
		filepath.Join(destination, j.dst),
		filepath.Join(destination, j.GetScriptSource(releaseBuild)),
	); err != nil {
		return err
	}
	return nil
}
