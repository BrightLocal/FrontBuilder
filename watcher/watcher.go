package watcher

import (
	"fmt"
	"log"

	"github.com/BrightLocal/FrontBuilder/builder"
	"github.com/fsnotify/fsnotify"
)

type BuildWatcher struct {
	Builder builder.FrontBuilder
	Watcher *fsnotify.Watcher
}

func NewBuildWatcher(builder builder.FrontBuilder) (*BuildWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &BuildWatcher{
		Builder: builder,
		Watcher: watcher,
	}, nil
}

func (bw *BuildWatcher) Watch() {
	fmt.Printf("Start watching files in directory %s\n", bw.Builder.GetFilesDir())
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = watcher.Close() }()
	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Printf("Rebuild file: %s", event.Name)
					bw.Builder.Build()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("got watch error: %s", err)
			}
		}
	}()

	for _, file := range bw.Builder.GetJSFiles() {
		err = watcher.Add(file)
		if err != nil {
			log.Fatal(err)
		}
	}
	for _, file := range bw.Builder.GetHTMLFiles() {
		err = watcher.Add(file)
		if err != nil {
			log.Fatal(err)
		}
	}
	for _, file := range bw.Builder.GetTSFiles() {
		err = watcher.Add(file)
		if err != nil {
			log.Fatal(err)
		}
	}
	<-done
}
