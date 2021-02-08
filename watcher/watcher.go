package watcher

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

type BuildWatcher struct {
	Watcher *fsnotify.Watcher
	paths   []string
}

func NewBuildWatcher(paths []string) (*BuildWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &BuildWatcher{
		Watcher: watcher,
		paths:   paths,
	}, nil
}

func (bw *BuildWatcher) Watch() (chan struct{}, error) {
	eventC := make(chan struct{})
	if err := bw.watchFolders(); err != nil {
		return nil, err
	}
	go func() {
		for event := range bw.Watcher.Events {
			if event.Op&fsnotify.Create != 0 && !strings.HasSuffix(event.Name, "~") {
				if err := bw.Watcher.Add(event.Name); err != nil {
					log.Printf("error add path %s to watch: %s", event.Name, err)
				}
			}
			if event.Op&fsnotify.Remove != 0 && !strings.HasSuffix(event.Name, "~") {
				if err := bw.Watcher.Remove(event.Name); err != nil {
					log.Printf("error remove path %s to watch: %s", event.Name, err)
				}
			}
			if event.Op&fsnotify.Chmod == 0 {
				eventC <- struct{}{}
			}
		}
	}()
	return eventC, nil
}

func (bw *BuildWatcher) watchFolders() error {
	for _, path := range bw.paths {
		if err := filepath.Walk(path, func(newPath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				log.Printf("Start watching files in directory: %s", newPath)
				err := bw.Watcher.Add(newPath)
				if err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}
