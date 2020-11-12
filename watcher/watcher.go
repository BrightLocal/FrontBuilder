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
	fmt.Printf("Start watching files in directory %s\n", bw.Builder.GetFilesDirectory())
	defer func() { _ = bw.Watcher.Close() }()
	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-bw.Watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Printf("Rebuild file: %s", event.Name)
					bw.Builder.Build()
				}
			case err, ok := <-bw.Watcher.Errors:
				if !ok {
					return
				}
				log.Printf("got watch error: %s", err)
			}
		}
	}()

	for _, file := range bw.Builder.GetJSFiles() {
		err := bw.Watcher.Add(file)
		if err != nil {
			log.Fatal(err)
		}
	}
	for _, file := range bw.Builder.GetHTMLFiles() {
		err := bw.Watcher.Add(file)
		if err != nil {
			log.Fatal(err)
		}
	}
	for _, file := range bw.Builder.GetTSFiles() {
		err := bw.Watcher.Add(file)
		if err != nil {
			log.Fatal(err)
		}
	}
	<-done
}
