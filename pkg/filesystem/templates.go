package filesystem

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/flowexec/flow/types/executable"
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
