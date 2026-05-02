package executable

import (
	"fmt"
	"regexp"

	"gopkg.in/yaml.v3"

	"github.com/flowexec/flow/v2/types/common"
)

//go:generate go run github.com/atombender/go-jsonschema@v0.16.0 -et --only-models -p executable -o flowfile.gen.go flowfile_schema.yaml

const FlowFileExt = ".flow"

var FlowFileExtRegex = regexp.MustCompile(fmt.Sprintf(`%s(\.yaml|\.yml)?$`, regexp.QuoteMeta(FlowFileExt)))

type FlowFileList []*FlowFile

func (f *FlowFile) SetContext(workspaceName, workspacePath, configPath string) {
	f.workspace = workspaceName
	f.workspacePath = workspacePath
	f.configPath = configPath
	for _, exec := range f.Executables {
		exec.SetContext(workspaceName, workspacePath, f.Namespace, configPath)
		if exec.Visibility == nil && f.Visibility != nil {
			v := ExecutableVisibility(*f.Visibility)
			exec.Visibility = &v
		}
		exec.SetDefaults()
		exec.SetInheritedFields(f)
	}
}

func (f *FlowFile) SetDefaults() {
	if f.Visibility == nil || *f.Visibility == "" {
		v := FlowFileVisibility(common.VisibilityPrivate)
		f.Visibility = &v
	}
}

func (f *FlowFile) WorkspacePath() string {
	return f.workspacePath
}

func (f *FlowFile) ConfigPath() string {
	return f.configPath
}

func (f *FlowFile) YAML() (string, error) {
	yamlBytes, err := yaml.Marshal(f)
	if err != nil {
		return "", fmt.Errorf("failed to marshal flowfile - %w", err)
	}
	return string(yamlBytes), nil
}

func (l *FlowFileList) FilterByNamespace(namespace string) FlowFileList {
	filteredCfgs := make(FlowFileList, 0)
	for _, cfg := range *l {
		if cfg.Namespace == namespace {
			filteredCfgs = append(filteredCfgs, cfg)
		}
	}
	return filteredCfgs
}

func (l *FlowFileList) FilterByTag(tag string) FlowFileList {
	filteredCfgs := make(FlowFileList, 0)
	for _, cfg := range *l {
		t := common.Tags(cfg.Tags)
		if t.HasTag(tag) {
			filteredCfgs = append(filteredCfgs, cfg)
		}
	}
	return filteredCfgs
}

func HasFlowFileExt(file string) bool {
	return FlowFileExtRegex.MatchString(file)
}
