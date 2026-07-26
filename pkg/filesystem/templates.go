package filesystem

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/flowexec/flow/v2/pkg/logger"
	"github.com/flowexec/flow/v2/types/executable"
	"github.com/flowexec/flow/v2/types/workspace"
)

func LoadFlowFileTemplate(flowfileName, templatePath string) (*executable.Template, error) {
	file, err := os.Open(filepath.Clean(templatePath))
	if err != nil {
		return nil, errors.Wrap(err, "unable to open template file")
	}
	defer file.Close()

	flowfileTmpl := &executable.Template{}
	err = yaml.NewDecoder(file).Decode(flowfileTmpl)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode template file")
	}
	flowfileTmpl.SetContext(flowfileName, templatePath)

	return flowfileTmpl, nil
}

func LoadFlowFileTemplates(templatePaths map[string]string) (executable.TemplateList, error) {
	templates := make(executable.TemplateList, 0, len(templatePaths))
	for name, path := range templatePaths {
		tmpl, err := LoadFlowFileTemplate(name, path)
		if err != nil {
			return nil, errors.Wrap(err, "unable to load flowfile templates")
		}
		templates = append(templates, tmpl)
	}
	return templates, nil
}

// LoadWorkspaceFlowFileTemplates discovers and loads every *.flow.tmpl file within the
// workspace, mirroring LoadWorkspaceFlowFiles for executables. Templates are named from
// their filename (via SetContext); files that fail to load are logged and skipped.
func LoadWorkspaceFlowFileTemplates(
	workspaceCfg *workspace.Workspace,
) (executable.TemplateList, error) {
	tmplFiles, err := findFlowFileTemplates(workspaceCfg)
	if err != nil {
		return nil, err
	}

	templates := make(executable.TemplateList, 0, len(tmplFiles))
	for _, tmplFile := range tmplFiles {
		tmpl, err := LoadFlowFileTemplate("", tmplFile)
		if err != nil {
			logger.Log().Error(
				fmt.Sprintf("unable to load flowfile template: %s", errors.Cause(err)),
				"file", tmplFile,
			)
			continue
		}
		templates = append(templates, tmpl)
	}
	logger.Log().Debug(
		fmt.Sprintf("loaded %d flowfile templates", len(templates)),
		"workspace",
		workspaceCfg.AssignedName(),
	)

	return templates, nil
}
