package builder

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/evanw/esbuild/pkg/api"
)

type Builder struct {
	source        string
	destination   string
	releaseBuild  bool
	indexFile     string
	htmlExtension string
	jsFiles       []string
	tsFiles       []string
	htmlFiles     []string
	jsApps        map[string]string
	buildOptions  []api.BuildOptions
	buildResult   []api.BuildResult
}

const (
	defaultIndex         = "index.html"
	defaultHTMLExtension = ".html"
)

type FrontBuilder interface {
	Build()
	GetSourceDirectory() string
}

func NewBuilder(source, destination string, releaseBuild bool) *Builder {
	return &Builder{
		source:        strings.TrimRight(source, "/") + "/",
		destination:   strings.TrimRight(destination, "/") + "/",
		releaseBuild:  releaseBuild,
		indexFile:     defaultIndex,
		htmlExtension: defaultHTMLExtension,
		jsApps:        make(map[string]string),
	}
}

func (b *Builder) IndexFile(fileName string) *Builder {
	b.indexFile = fileName
	return b
}

func (b *Builder) HTMLExtension(ext string) *Builder {
	b.htmlExtension = "." + strings.TrimLeft(ext, ".")
	return b
}

func (b *Builder) Build() error {
	start := time.Now()
	if err := b.clearStaticDir(); err != nil {
		return fmt.Errorf("error clear destination folder: %s", err)
	}
	if err := b.collectFiles(); err != nil {
		return fmt.Errorf("error collect source files: %s", err)
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
		if err := b.checkBuildErrors(); err != nil {
			return fmt.Errorf("build failed: %s", err)
		}
		if err := b.processJSFiles(); err != nil {
			return fmt.Errorf("fail process js files: %s", err)
		}
		fmt.Printf("JS processed in %s\n", time.Since(start))
	}
	{
		start := time.Now()
		if err := b.processHTMLFiles(); err != nil {
			return fmt.Errorf("fail process HTML files: %s", err)
		}
		fmt.Printf("HTML processed in %s\n", time.Since(start))
	}
	fmt.Printf("Build completed in %s\n", time.Since(start))
	return nil
}

func (b *Builder) GetSourceDirectory() string {
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
			b.jsFiles = append(b.jsFiles, strings.TrimPrefix(path, b.source))
		case strings.HasSuffix(info.Name(), b.htmlExtension):
			b.htmlFiles = append(b.htmlFiles, strings.TrimPrefix(path, b.source))
		case strings.HasSuffix(info.Name(), ".ts"):
			b.tsFiles = append(b.tsFiles, strings.TrimPrefix(path, b.source))
		}
		return nil
	})
}

func (b *Builder) prepareApps() {
	for _, file := range b.jsFiles {
		if b.contains(b.htmlFiles, file) {
			b.jsApps[strings.TrimSuffix(file, ".js")+b.htmlExtension] = file
		}
	}
	for _, file := range b.tsFiles {
		if b.contains(b.htmlFiles, file) {
			b.jsApps[strings.TrimSuffix(file, ".ts")+b.htmlExtension] = file
		}
	}
}

func (b *Builder) getDefaultBuildOption() api.BuildOptions {
	if b.releaseBuild {
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
		buildOption := b.getDefaultBuildOption()
		buildOption.Outdir = filepath.Join(b.destination, "/js", filepath.Dir(jsFile))
		if !strings.HasSuffix(jsFile, ".ts") {
			buildOption.EntryPoints = []string{filepath.Join(b.source, jsFile)}
			buildOptions = append(buildOptions, buildOption)
		} else {
			buildOption.EntryPoints = []string{filepath.Join(b.source, jsFile)}
			buildOption.Loader = map[string]api.Loader{".ts": api.LoaderTS}
			buildOption.Tsconfig = "tsconfig.json"
			buildOptions = append(buildOptions, buildOption)
		}
	}
	b.buildOptions = buildOptions
}

func (b *Builder) processHTMLFiles() error {
	for _, htmlFile := range b.htmlFiles {
		if err := b.prepareHTMLFile(htmlFile); err != nil {
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

func (b *Builder) checkBuildErrors() error {
	for _, result := range b.buildResult {
		if len(result.Errors) > 0 {
			for _, err := range result.Errors {
				fmt.Printf("Error in %s:%d: %s\n", err.Location.File, err.Location.Line, err.Text)
			}
			return errors.New("errors on build process. check above messages")
		}
	}
	return nil
}

func (b *Builder) processJSFiles() error {
	for _, result := range b.buildResult {
		for _, file := range result.OutputFiles {
			dir, fileName := filepath.Split(file.Path)
			if b.releaseBuild {
				if !strings.HasSuffix(file.Path, ".map") {
					hashSum := md5.Sum(file.Contents)
					fileHash := fmt.Sprintf("%x", hashSum)[:8]
					ext := path.Ext(fileName)
					outfile := dir + fileName[0:len(fileName)-len(ext)] + "." + fileHash + ".js"
					if err := os.Rename(file.Path, outfile); err != nil {
						return err
					}
					compiledFile := strings.TrimPrefix(file.Path, filepath.Join(b.destination, "/js")+"/")
					htmlFile := strings.TrimSuffix(compiledFile, ".js") + b.htmlExtension
					if val, ok := b.jsApps[htmlFile]; ok && val != "" {
						b.jsApps[htmlFile] = outfile
					}
				}
			} else {
				if !strings.HasSuffix(file.Path, ".map") {
					compiledFile := strings.TrimPrefix(file.Path, filepath.Join(b.destination, "/js")+"/")
					htmlFile := strings.TrimSuffix(compiledFile, ".js") + b.htmlExtension
					if val, ok := b.jsApps[htmlFile]; ok && val != "" {
						b.jsApps[htmlFile] = strings.TrimPrefix(file.Path, b.destination)
					}
				}
			}
		}
	}
	return nil
}

func (b *Builder) prepareHTMLFile(htmlFile string) error {
	htmlSrc, err := ioutil.ReadFile(filepath.Join(b.source, htmlFile))
	if err != nil && err != io.EOF {
		return err
	}
	if jsFile, ok := b.jsApps[htmlFile]; ok {
		filename := strings.TrimPrefix(jsFile, b.destination)
		if filepath.Base(htmlFile) == b.indexFile {
			scriptTag := []byte(`<script src=""></script>`)
			if bytes.Contains(htmlSrc, scriptTag) {
				htmlSrc = bytes.Replace(htmlSrc, []byte(`src=""`), []byte(`src="/`+filename+`"`), -1)
			}
		} else {
			scriptTag := []byte("\n{{ define \"js-app\" }}<script src=\"/" + filename + "\"></script>{{ end }}\n")
			htmlSrc = append(htmlSrc, scriptTag...)
		}
	}
	dst := filepath.Join(b.destination, "views", htmlFile)
	if err := os.MkdirAll(filepath.Dir(dst), 0770); err != nil {
		return err
	}
	if err := ioutil.WriteFile(dst, htmlSrc, 0644); err != nil {
		return err
	}
	return nil
}

func (b Builder) contains(s []string, e string) bool {
	for _, htmlFile := range s {
		if htmlPath := strings.TrimSuffix(htmlFile, b.htmlExtension); htmlPath == strings.TrimSuffix(e, ".js") ||
			htmlPath == strings.TrimSuffix(e, ".ts") {
			return true
		}
	}
	return false
}
