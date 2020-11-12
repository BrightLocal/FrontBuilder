package watcher

import (
	"errors"
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
)

type BuildWatcher struct {
	Watcher *fsnotify.Watcher
}

func NewBuildWatcher() (*BuildWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &BuildWatcher{
		Watcher: watcher,
	}, nil
}

func (bw *BuildWatcher) Watch(directory string) (chan struct{}, error) {
	if stat, err := os.Stat(directory); err != nil {
		return nil, err
	} else if !stat.IsDir() {
		return nil, errors.New("not a directory")
	}
	eventC := make(chan struct{})
	fmt.Printf("Start watching files in directory %s\n", directory)
	go func() {
		for event := range bw.Watcher.Events {
			if event.Op&fsnotify.Chmod == 0 {
				eventC <- struct{}{}
			}
		}
	}()
	err := bw.Watcher.Add(directory)
	if err != nil {
		return nil, err
	}
	return eventC, nil
}
