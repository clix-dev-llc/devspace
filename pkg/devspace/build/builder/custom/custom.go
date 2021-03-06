package custom

import (
	"io"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar"
	"github.com/loft-sh/devspace/pkg/devspace/config/generated"
	"github.com/loft-sh/devspace/pkg/devspace/config/versions/latest"
	"github.com/loft-sh/devspace/pkg/util/command"
	"github.com/loft-sh/devspace/pkg/util/hash"
	logpkg "github.com/loft-sh/devspace/pkg/util/log"

	dockerterm "github.com/docker/docker/pkg/term"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

var (
	_, stdout, stderr = dockerterm.StdStreams()
)

// Builder holds all the relevant information for a custom build
type Builder struct {
	imageConf *latest.ImageConfig

	imageConfigName string
	imageTags       []string
}

// NewBuilder creates a new custom builder
func NewBuilder(imageConfigName string, imageConf *latest.ImageConfig, imageTags []string) *Builder {
	return &Builder{
		imageConfigName: imageConfigName,
		imageConf:       imageConf,
		imageTags:       imageTags,
	}
}

// ShouldRebuild implements interface
func (b *Builder) ShouldRebuild(cache *generated.CacheConfig, forceRebuild, ignoreContextPathChanges bool) (bool, error) {
	if b.imageConf.Build.Custom.OnChange == nil || len(b.imageConf.Build.Custom.OnChange) == 0 {
		return true, nil
	}

	// Hash image config
	configStr, err := yaml.Marshal(*b.imageConf)
	if err != nil {
		return false, errors.Wrap(err, "marshal image config")
	}
	imageConfigHash := hash.String(string(configStr))

	// Loop over on change globs
	customFilesHash := ""
	for _, pattern := range b.imageConf.Build.Custom.OnChange {
		files, err := doublestar.Glob(pattern)
		if err != nil {
			return false, err
		}

		for _, file := range files {
			sha256, err := hash.Directory(file)
			if err != nil {
				return false, errors.Wrap(err, "hash "+file)
			}

			customFilesHash += sha256
		}
	}
	customFilesHash = hash.String(customFilesHash)

	imageCache := cache.GetImageCache(b.imageConfigName)

	// only rebuild Docker image when Dockerfile or context has changed since latest build
	mustRebuild := imageCache.Tag == "" || imageCache.ImageConfigHash != imageConfigHash || imageCache.CustomFilesHash != customFilesHash

	imageCache.ImageConfigHash = imageConfigHash
	imageCache.CustomFilesHash = customFilesHash

	return mustRebuild, nil
}

// Build implements interface
func (b *Builder) Build(log logpkg.Logger) error {
	// Build arguments
	args := []string{}

	// add args
	for _, arg := range b.imageConf.Build.Custom.Args {
		args = append(args, arg)
	}

	// add tags
	for _, tag := range b.imageTags {
		if b.imageConf.Build.Custom.ImageFlag != "" {
			args = append(args, b.imageConf.Build.Custom.ImageFlag)
		}

		args = append(args, b.imageConf.Image+":"+tag)
	}

	for _, arg := range b.imageConf.Build.Custom.AppendArgs {
		args = append(args, arg)
	}

	// Create the command
	cmd := command.NewStreamCommand(filepath.FromSlash(b.imageConf.Build.Custom.Command), args)

	// Determine output writer
	var writer io.Writer
	if log == logpkg.GetInstance() {
		writer = stdout
	} else {
		writer = log
	}

	log.Infof("Build %s:%s with custom command %s %s", b.imageConf.Image, b.imageTags[0], b.imageConf.Build.Custom.Command, strings.Join(args, " "))

	err := cmd.Run(writer, writer, nil)
	if err != nil {
		return errors.Errorf("Error building image: %v", err)
	}

	log.Done("Done processing image '" + b.imageConf.Image + "'")
	return nil
}
