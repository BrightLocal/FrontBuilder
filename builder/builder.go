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

type sourcePath struct {
	BaseDir string
	Path    string
}

type Builder struct {
	sources          []string
	destination      string
	releaseBuild     bool
	indexFile        string
	htmlExtension    string
	scriptsPrefix    string
	htmlPrefix       string
	typeScriptConfig string
	scripts          map[string]sourcePath
	typeScripts      map[string]sourcePath
	htmls            map[string]*files.HTML
	jsApps           map[string]sourcePath
	buildOptions     []api.BuildOptions
	buildResult      []api.BuildResult
}

const (
	defaultIndexFile        = "index.html"
	defaultHTMLExtension    = ".html"
	defaultTypeScriptConfig = "tsconfig.json"
)

type FrontBuilder interface {
	Build()
	GetSourceDirectory() string
}

func NewBuilder(sources []string, destination string, releaseBuild bool) *Builder {
	return &Builder{
		sources:          sources,
		destination:      strings.TrimRight(destination, "/") + "/",
		releaseBuild:     releaseBuild,
		indexFile:        defaultIndexFile,
		htmlExtension:    defaultHTMLExtension,
		typeScriptConfig: defaultTypeScriptConfig,
		jsApps:           make(map[string]sourcePath),
		htmls:            make(map[string]*files.HTML),
		scripts:          make(map[string]sourcePath),
		typeScripts:      make(map[string]sourcePath),
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

func (b *Builder) ScriptsPrefix(scriptsPrefix string) *Builder {
	b.scriptsPrefix = scriptsPrefix
	return b
}

func (b *Builder) HTMLPrefix(htmlPrefix string) *Builder {
	b.htmlPrefix = htmlPrefix
	return b
}

func (b *Builder) TypeScriptConfig(path string) *Builder {
	b.typeScriptConfig = path
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
	for _, source := range b.sources {
		if err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			if err := b.collectFileTypes(info, source, path); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) collectFileTypes(fileInfo os.FileInfo, source, path string) error {
	switch {
	case strings.HasSuffix(fileInfo.Name(), ".js"):
		b.scripts[fileInfo.Name()] = sourcePath{
			BaseDir: source,
			Path:    strings.TrimPrefix(path, source),
		}
	case strings.HasSuffix(fileInfo.Name(), ".ts"):
		b.typeScripts[fileInfo.Name()] = sourcePath{
			BaseDir: source,
			Path:    strings.TrimPrefix(path, source),
		}
	case strings.HasSuffix(fileInfo.Name(), b.htmlExtension):
		name := strings.TrimPrefix(path, source)
		if _, ok := b.htmls[name]; ok {
			if b.releaseBuild {
				return errors.New("duplicate source: " + name)
			}
		}
		b.htmls[name] = files.NewHTML(filepath.Join(source, strings.TrimPrefix(path, source)))
	}
	return nil
}

func (b *Builder) prepareApps() {
	for _, script := range b.scripts {
		html := strings.TrimSuffix(script.Path, ".js") + b.htmlExtension
		if _, ok := b.htmls[html]; ok {
			b.jsApps[html] = script
		}
	}
	for _, script := range b.typeScripts {
		html := strings.TrimSuffix(script.Path, ".ts") + b.htmlExtension
		if _, ok := b.htmls[html]; ok {
			b.jsApps[html] = script
		}
	}
}

func (b *Builder) prepareBuildOptions() {
	var buildOptions []api.BuildOptions
	b.buildOptions = []api.BuildOptions{}
	for _, jsFile := range b.jsApps {
		buildOption := b.getDefaultBuildOption()
		buildOption.Outdir = filepath.Join(b.destination, b.scriptsPrefix, filepath.Dir(jsFile.Path))
		buildOption.EntryPoints = []string{filepath.Join(jsFile.BaseDir, jsFile.Path)}
		if strings.HasSuffix(jsFile.Path, ".ts") {
			buildOption.Loader = map[string]api.Loader{".ts": api.LoaderTS}
			buildOption.Tsconfig = b.typeScriptConfig
		}
		buildOptions = append(buildOptions, buildOption)
	}
	b.buildOptions = buildOptions
}

func (b *Builder) build() {
	b.buildResult = []api.BuildResult{}
	for _, buildOption := range b.buildOptions {
		b.buildResult = append(b.buildResult, api.Build(buildOption))
	}
}

func (b *Builder) checkBuildErrors() error {
	for _, result := range b.buildResult {
		if len(result.Errors) > 0 {
			for _, err := range result.Errors {
				if err.Location != nil {
					fmt.Printf("Error in %s:%d: %s\n", err.Location.File, err.Location.Line, err.Text)
				}
			}
			return errors.New("errors on build process. check above messages")
		}
	}
	return nil
}

func (b *Builder) processHTMLFiles() error {
	resultFiles := b.resultFiles()
	for path, html := range b.htmls {
		script := strings.TrimSuffix(filepath.Join(b.destination, b.scriptsPrefix, path), b.htmlExtension) + ".js"
		if content, ok := resultFiles[script]; ok {
			html.InjectJS(files.NewJS(b.destination, script, content))
		}
		if err := html.Render(filepath.Join(b.destination, b.htmlPrefix, path), b.releaseBuild); err != nil {
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
