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
) error {
	var errs []error
	for i, a := range artifacts {
		if err := copyArtifact(
			fmt.Sprintf("artifact-%d", i), wsDir, srcDir, dstDir, a, templateData,
		); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Errorf("errors copying artifacts: %v", errs)
	}
	return nil
}

//nolint:gocognit
func copyArtifact(
	name, wsPath, srcDir, dstDir string,
	artifact executable.Artifact,
	templateData expressionData,
) error {
	srcPath, err := parseSourcePath(name, srcDir, wsPath, artifact, templateData)
	if err != nil {
		return errors.Wrap(err, "unable to parse source path")
	}

	if artifact.If != "" {
		eval, err := expression.IsTruthy(artifact.If, templateData)
		if err != nil {
			return errors.Wrap(err, "unable to evaluate if condition")
		}
		if !eval {
			logger.Log().Debugf("skipping artifact %s", name)
			return nil
		}
	}

	srcName := filepath.Base(srcPath)
	if strings.Contains(srcName, "*") {
		matches, err := filepath.Glob(srcPath)
		if err != nil {
			return errors.Wrap(err, "unable to glob source path")
		}
		var errs []error
		for i, match := range matches {
			m := artifact
			m.SrcName = filepath.Base(match)
			m.SrcDir = filepath.Dir(match)
			mErr := copyArtifact(fmt.Sprintf("%s-%d", name, i), wsPath, srcDir, dstDir, m, templateData)
			if mErr != nil {
				errs = append(errs, mErr)
			}
		}
		if len(errs) > 0 {
			return errors.Errorf("errors copying artifact from pattern: %v", errs)
		}
	}

	info, err := os.Stat(srcPath)
	switch {
	case os.IsNotExist(err):
		return errors.Errorf("file does not exist: %s", srcPath)
	case err != nil:
		return errors.Wrap(err, "unable to stat src file")
	case info.IsDir():
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
			return copyArtifact(aName, wsPath, srcDir, dstDir, a, templateData)
		})
		if err != nil {
			return errors.Wrap(err, "unable to walk directory")
		}
	}
	if artifact.DstName == "" {
		artifact.DstName = srcName
	}
	dstPath, err := parseDestinationPath(
		name,
		dstDir, srcDir, wsPath,
		artifact,
		templateData,
	)
	if err != nil {
		return errors.Wrap(err, "unable to parse destination path")
	}

	if err := os.MkdirAll(dstDir, 0750); err != nil {
		if !os.IsExist(err) {
			return errors.Wrap(err, "unable to create destination directory")
		}
		return errors.Wrap(err, "unable to create destination directory")
	}

	logger.Log().Debugx("copying artifact", "name", name, "src", srcPath, "dst", dstPath)
	if _, e := os.Stat(dstPath); e == nil {
		// TODO: Add a flag to overwrite existing files
		logger.Log().Warnx("Overwriting existing file", "dst", dstPath)
	}
	if err := filesystem.CopyFile(srcPath, dstPath); err != nil {
		return errors.Wrap(err, "unable to copy artifact")
	}
	return nil
}
