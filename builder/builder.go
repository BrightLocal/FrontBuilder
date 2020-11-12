package builder

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/evanw/esbuild/pkg/api"
)

type Builder struct {
	source       string
	destination  string
	environment  string
	jsFiles      []string
	tsFiles      []string
	htmlFiles    []string
	jsApps       map[string]string
	buildOptions []api.BuildOptions
	buildResult  []api.BuildResult
}

type FrontBuilder interface {
	Build()
	GetJSFiles() []string
	GetHTMLFiles() []string
	GetTSFiles() []string
	GetFilesDirectory() string
}

func NewBuilder(source, destination, env string) *Builder {
	return &Builder{
		source:      source,
		destination: destination,
		environment: env,
		jsApps:      make(map[string]string),
	}
}

func (b *Builder) Build() {
	start := time.Now()
	if err := b.clearStaticDir(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := b.collectFiles(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("%d files collected in %s\n",
		len(b.htmlFiles)+len(b.tsFiles)+len(b.jsFiles),
		time.Since(start),
	)
	{
		start := time.Now()
		b.prepareApps()
		b.prepareBuildOptions()
		fmt.Printf("Apps prepared in %s\n", time.Since(start))
	}
	{
		start := time.Now()
		b.build()
		b.checkBuildErrors()
		if err := b.processJSFiles(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("JS processed in %s\n", time.Since(start))
	}
	{
		start := time.Now()
		if err := b.processHTMLFiles(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("HTML processed in %s\n", time.Since(start))
	}
	fmt.Printf("Build completed in %s\n", time.Since(start))
}

func (b *Builder) GetJSFiles() []string {
	return b.jsFiles
}

func (b *Builder) GetHTMLFiles() []string {
	return b.htmlFiles
}

func (b *Builder) GetTSFiles() []string {
	return b.tsFiles
}

func (b *Builder) GetFilesDirectory() string {
	return b.source
}

func (b *Builder) clearStaticDir() error {
	dir, err := ioutil.ReadDir(b.destination)
	if err != nil {
		return err
	}
	for _, d := range dir {
		if d.IsDir() && (d.Name() == "views" || d.Name() == "js") {
			if err := os.RemoveAll(filepath.Join(b.destination, d.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *Builder) collectFiles() error {
	return filepath.Walk(b.source, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		switch {
		case strings.HasSuffix(info.Name(), ".js"):
			b.jsFiles = append(b.jsFiles, path)
		case strings.HasSuffix(info.Name(), ".html"):
			b.htmlFiles = append(b.htmlFiles, path)
		case strings.HasSuffix(info.Name(), ".ts"):
			b.tsFiles = append(b.tsFiles, path)
		}
		return nil
	})
}

func (b *Builder) prepareApps() {
	for _, file := range b.jsFiles {
		if contains(b.htmlFiles, file) {
			b.jsApps[strings.TrimSuffix(file, ".js")+".html"] = file
		}
	}
	for _, file := range b.tsFiles {
		if contains(b.htmlFiles, file) {
			b.jsApps[strings.TrimSuffix(file, ".ts")+".html"] = file
		}
	}
}

func (b *Builder) getDefaultBuildOption() api.BuildOptions {
	if b.environment == "production" {
		return api.BuildOptions{
			Bundle:            true,
			Write:             true,
			LogLevel:          api.LogLevelInfo,
			Sourcemap:         api.SourceMapLinked,
			Target:            api.ESNext,
			MinifyWhitespace:  true,
			MinifyIdentifiers: true,
			MinifySyntax:      true,
		}
	}
	return api.BuildOptions{
		Bundle:    true,
		Write:     true,
		LogLevel:  api.LogLevelInfo,
		Sourcemap: api.SourceMapNone,
		Target:    api.ESNext,
	}
}

func (b *Builder) prepareBuildOptions() {
	var buildOptions []api.BuildOptions
	for _, jsFile := range b.jsApps {
		dir, _ := filepath.Split(jsFile)
		paths := strings.Split(strings.TrimSuffix(dir, "/"), "/")
		folder := strings.Join(paths[1:], "/")
		buildOption := b.getDefaultBuildOption()
		buildOption.Outdir = filepath.Join(b.destination, "/js", folder)
		if !strings.HasSuffix(jsFile, ".ts") {
			buildOption.EntryPoints = []string{jsFile}
			buildOptions = append(buildOptions, buildOption)
		} else {
			buildOption.EntryPoints = []string{jsFile}
			buildOption.Loader = map[string]api.Loader{".ts": api.LoaderTS}
			buildOption.Tsconfig = "tsconfig.json"
			buildOptions = append(buildOptions, buildOption)
		}
	}
	b.buildOptions = buildOptions
}

func (b *Builder) processHTMLFiles() error {
	for _, htmlFile := range b.htmlFiles {
		dir, filename := filepath.Split(htmlFile)
		paths := strings.Split(strings.TrimSuffix(dir, "/"), "/")
		folder := strings.Join(paths[1:], "/")
		destPath := filepath.Join(b.destination, "views", folder, filename)
		if err := b.prepareHTMLFile(htmlFile, destPath); err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) build() {
	for _, buildOption := range b.buildOptions {
		result := api.Build(buildOption)
		b.buildResult = append(b.buildResult, result)
	}
}

func (b *Builder) checkBuildErrors() {
	for _, result := range b.buildResult {
		if len(result.Errors) > 0 {
			for _, err := range result.Errors {
				fmt.Printf("Error in %s:%d: %s\n", err.Location.File, err.Location.Line, err.Text)
			}
			os.Exit(1)
		}
	}
}

func (b *Builder) processJSFiles() error {
	for _, result := range b.buildResult {
		for _, file := range result.OutputFiles {
			dir, fileName := filepath.Split(file.Path)
			if b.environment == "production" {
				if !strings.HasSuffix(file.Path, ".map") {
					hashSum := md5.Sum(file.Contents)
					fileHash := fmt.Sprintf("%x", hashSum)[:8]
					ext := path.Ext(fileName)
					outfile := dir + fileName[0:len(fileName)-len(ext)] + "." + fileHash + ".js"
					if err := os.Rename(file.Path, outfile); err != nil {
						return err
					}
					compiledFile := strings.TrimPrefix(file.Path, filepath.Join(b.destination, "/js"))
					dir := strings.TrimPrefix(b.source, "./")
					htmlFile := strings.TrimSuffix(compiledFile, ".js") + ".html"
					log.Printf("html files-> %s", htmlFile)
					if val, ok := b.jsApps[dir+htmlFile]; ok && val != "" {
						b.jsApps[dir+htmlFile] = outfile
					}
				}
			} else {
				if !strings.HasSuffix(file.Path, ".map") {
					compiledFile := strings.TrimPrefix(file.Path, filepath.Join(b.destination, "/js"))
					dir := strings.TrimPrefix(b.source, "./")
					htmlFile := strings.TrimSuffix(compiledFile, ".js") + ".html"
					if val, ok := b.jsApps[dir+htmlFile]; ok && val != "" {
						b.jsApps[dir+htmlFile] = file.Path
					}
				}
			}
		}
	}
	return nil
}

func (b *Builder) prepareHTMLFile(src, dst string) error {
	data, err := ioutil.ReadFile(src)
	if err != nil && err != io.EOF {
		return err
	}
	if jsFile, ok := b.jsApps[src]; ok {
		filename := strings.TrimPrefix(jsFile, b.destination)
		// @TODO replace hardcoded index.html
		if strings.HasSuffix(src, "index.html") {
			scriptTag := []byte(`<script src=""></script>`)
			if bytes.Contains(data, scriptTag) {
				data = bytes.Replace(data, []byte(`src=""`), []byte(`src="`+filename+`"`), -1)
			}
		} else {
			scriptTag := []byte("\n{{ define \"js-app\" }}<script src=\"" + filename + "\"></script>{{ end }}\n")
			data = append(data, scriptTag...)
		}
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0770); err != nil {
		return err
	}
	if err := ioutil.WriteFile(dst, data, 0644); err != nil {
		return err
	}
	return nil
}

func contains(s []string, e string) bool {
	for _, htmlFile := range s {
		if htmlPath := strings.TrimSuffix(htmlFile, ".html"); htmlPath == strings.TrimSuffix(e, ".js") ||
			htmlPath == strings.TrimSuffix(e, ".ts") {
			return true
		}
	}
	return false
}
