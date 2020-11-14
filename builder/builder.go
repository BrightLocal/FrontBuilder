package builder

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BrightLocal/FrontBuilder/builder/files"
	"github.com/evanw/esbuild/pkg/api"
)

type Builder struct {
	source        string
	destination   string
	releaseBuild  bool
	indexFile     string
	htmlExtension string
	scripts       []string
	typeScripts   []string
	htmls         map[string]*files.HTML
	jsApps        map[string]string
	buildOptions  []api.BuildOptions
	buildResult   []api.BuildResult
}

const (
	defaultIndexFile     = "index.html"
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
		indexFile:     defaultIndexFile,
		htmlExtension: defaultHTMLExtension,
		jsApps:        make(map[string]string),
		htmls:         make(map[string]*files.HTML),
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
	if err := b.collectFiles(); err != nil {
		return fmt.Errorf("error collecting files: %s", err)
	}
	b.prepareApps()
	b.prepareBuildOptions()
	b.build()
	if err := b.checkBuildErrors(); err != nil {
		return fmt.Errorf("build failed: %s", err)
	}
	if err := b.processHTMLFiles(); err != nil {
		return fmt.Errorf("error processing HTMLs: %s", err)
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
			b.scripts = append(b.scripts, strings.TrimPrefix(path, b.source))
		case strings.HasSuffix(info.Name(), ".ts"):
			b.typeScripts = append(b.typeScripts, strings.TrimPrefix(path, b.source))
		case strings.HasSuffix(info.Name(), b.htmlExtension):
			b.htmls[info.Name()] = files.NewHTML(info.Name())
		}
		return nil
	})
}

func (b *Builder) prepareApps() {
	for _, script := range b.scripts {
		html := strings.TrimSuffix(script, ".js") + b.htmlExtension
		if _, ok := b.htmls[html]; ok {
			b.jsApps[html] = script
		}
	}
	for _, script := range b.typeScripts {
		html := strings.TrimSuffix(script, ".ts") + b.htmlExtension
		if _, ok := b.htmls[html]; ok {
			b.jsApps[html] = script
		}
	}
}

func (b *Builder) prepareBuildOptions() {
	var buildOptions []api.BuildOptions
	for _, jsFile := range b.jsApps {
		buildOption := b.getDefaultBuildOption()
		buildOption.Outdir = filepath.Join(b.destination, filepath.Dir(jsFile))
		buildOption.EntryPoints = []string{filepath.Join(b.source, jsFile)}
		if strings.HasSuffix(jsFile, ".ts") {
			buildOption.Loader = map[string]api.Loader{".ts": api.LoaderTS}
			buildOption.Tsconfig = "tsconfig.json"
		}
		buildOptions = append(buildOptions, buildOption)
	}
	b.buildOptions = buildOptions
}

func (b *Builder) build() {
	for _, buildOption := range b.buildOptions {
		b.buildResult = append(b.buildResult, api.Build(buildOption))
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

func (b *Builder) processHTMLFiles() error {
	resultFiles := b.resultFiles()
	for path, html := range b.htmls {
		script := strings.TrimSuffix(path, b.htmlExtension) + ".js"
		if content, ok := resultFiles[script]; ok {
			html.InjectJS(files.NewJS(script, content))
		}
		if err := html.Render(filepath.Join(b.destination, path), b.releaseBuild); err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) getDefaultBuildOption() api.BuildOptions {
	if b.releaseBuild {
		return releaseBuildOptions
	}
	return devBuildOptions
}

func (b *Builder) resultFiles() map[string][]byte {
	htmlScripts := make(map[string][]byte)
	for _, result := range b.buildResult {
		for _, file := range result.OutputFiles {
			htmlScripts[file.Path] = file.Contents
		}
	}
	return htmlScripts
}

/*
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
*/
/*
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
*/
