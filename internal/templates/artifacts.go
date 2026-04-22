package templates

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/jahvon/expression"
	"github.com/pkg/errors"

	"github.com/flowexec/flow/pkg/filesystem"
	"github.com/flowexec/flow/pkg/logger"
	"github.com/flowexec/flow/types/executable"
)

func copyAllArtifacts(
	artifacts []executable.Artifact,
	wsDir, srcDir, dstDir string,
	templateData expressionData,
) ([]string, error) {
	var copied []string
	var errs []error
	for i, a := range artifacts {
		paths, err := copyArtifact(
			fmt.Sprintf("artifact-%d", i), wsDir, srcDir, dstDir, a, templateData,
		)
		if err != nil {
			errs = append(errs, err)
		}
		copied = append(copied, paths...)
	}
	if len(errs) > 0 {
		return copied, errors.Errorf("errors copying artifacts: %v", errs)
	}
	return copied, nil
}

func copyArtifact(
	name, wsPath, srcDir, dstDir string,
	artifact executable.Artifact,
	templateData expressionData,
) ([]string, error) {
	srcPath, err := parseSourcePath(name, srcDir, wsPath, artifact, templateData)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse source path")
	}

	if artifact.If != "" {
		eval, err := expression.IsTruthy(artifact.If, templateData)
		if err != nil {
			return nil, errors.Wrap(err, "unable to evaluate if condition")
		}
		if !eval {
			logger.Log().Debugf("skipping artifact %s", name)
			return nil, nil
		}
	}

	srcName := filepath.Base(srcPath)
	if strings.Contains(srcName, "*") {
		return copyGlobArtifacts(name, wsPath, srcDir, dstDir, srcPath, artifact, templateData)
	}

	info, err := os.Stat(srcPath)
	switch {
	case os.IsNotExist(err):
		return nil, errors.Errorf("file does not exist: %s", srcPath)
	case err != nil:
		return nil, errors.Wrap(err, "unable to stat src file")
	case info.IsDir():
		return copyDirArtifacts(name, wsPath, srcDir, dstDir, srcPath, artifact, templateData)
	}

	if artifact.DstName == "" {
		artifact.DstName = srcName
	}
	dstPath, err := parseDestinationPath(name, dstDir, srcDir, wsPath, artifact, templateData)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse destination path")
	}

	if err := os.MkdirAll(dstDir, 0750); err != nil {
		return nil, errors.Wrap(err, "unable to create destination directory")
	}

	logger.Log().Debug("copying artifact", "name", name, "src", srcPath, "dst", dstPath)
	if _, e := os.Stat(dstPath); e == nil {
		// TODO: Add a flag to overwrite existing files
		logger.Log().Warn("Overwriting existing file", "dst", dstPath)
	}
	if err := filesystem.CopyFile(srcPath, dstPath); err != nil {
		return nil, errors.Wrap(err, "unable to copy artifact")
	}
	return []string{dstPath}, nil
}

func copyGlobArtifacts(
	name, wsPath, srcDir, dstDir, srcPath string,
	artifact executable.Artifact,
	templateData expressionData,
) ([]string, error) {
	matches, err := filepath.Glob(srcPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to glob source path")
	}
	var copied []string
	var errs []error
	for i, match := range matches {
		m := artifact
		m.SrcName = filepath.Base(match)
		m.SrcDir = filepath.Dir(match)
		paths, mErr := copyArtifact(fmt.Sprintf("%s-%d", name, i), wsPath, srcDir, dstDir, m, templateData)
		if mErr != nil {
			errs = append(errs, mErr)
		}
		copied = append(copied, paths...)
	}
	if len(errs) > 0 {
		return copied, errors.Errorf("errors copying artifact from pattern: %v", errs)
	}
	return copied, nil
}

func copyDirArtifacts(
	name, wsPath, srcDir, dstDir, srcPath string,
	artifact executable.Artifact,
	templateData expressionData,
) ([]string, error) {
	var copied []string
	err := filepath.WalkDir(srcPath, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		a := artifact
		a.SrcName = filepath.Base(path)
		a.SrcDir = filepath.Dir(path)
		aName := fmt.Sprintf("%s-%s", name, a.SrcName)
		paths, walkErr := copyArtifact(aName, wsPath, srcDir, dstDir, a, templateData)
		copied = append(copied, paths...)
		return walkErr
	})
	if err != nil {
		return copied, errors.Wrap(err, "unable to walk directory")
	}
	return copied, nil
}
