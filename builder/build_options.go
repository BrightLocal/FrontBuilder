package builder

import "github.com/evanw/esbuild/pkg/api"

var (
	releaseBuildOptions = api.BuildOptions{
		Bundle:            true,
		Write:             true,
		LogLevel:          api.LogLevelInfo,
		Sourcemap:         api.SourceMapLinked,
		Target:            api.ESNext,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
	}
	devBuildOptions = api.BuildOptions{
		Bundle:    true,
		Write:     true,
		LogLevel:  api.LogLevelInfo,
		Sourcemap: api.SourceMapNone,
		Target:    api.ESNext,
	}
)
