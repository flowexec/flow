package executable

import (
	"github.com/flowexec/flow/internal/io/common"
	"github.com/flowexec/flow/internal/logger"
	"github.com/flowexec/flow/types/executable"
)

func PrintExecutableList(format string, executables executable.ExecutableList) {
	logger.Log().Debugf("listing %d executables", len(executables))
	switch common.NormalizeFormat(format) {
	case common.YAMLFormat:
		str, err := executables.YAML()
		if err != nil {
			logger.Log().Fatalf("Failed to marshal executable list - %v", err)
		}
		logger.Log().Println(str)
	case common.JSONFormat:
		str, err := executables.JSON()
		if err != nil {
			logger.Log().Fatalf("Failed to marshal executable list - %v", err)
		}
		logger.Log().Println(str)
	default:
		logger.Log().Fatalf("Unsupported output format %s", format)
	}
}

func PrintExecutable(format string, exec *executable.Executable) {
	if exec == nil {
		logger.Log().Fatalf("Executable type is nil")
	}
	switch common.NormalizeFormat(format) {
	case common.YAMLFormat:
		str, err := exec.YAML()
		if err != nil {
			logger.Log().Fatalf("Failed to marshal executable - %v", err)
		}
		logger.Log().Println(str)
	case common.JSONFormat:
		str, err := exec.JSON()
		if err != nil {
			logger.Log().Fatalf("Failed to marshal executable - %v", err)
		}
		logger.Log().Println(str)
	}
}

func PrintTemplate(format string, template *executable.Template) {
	if template == nil {
		logger.Log().Fatalf("Template type is nil")
	}
	logger.Log().Debugf("Template %s", template.Name())
	switch common.NormalizeFormat(format) {
	case common.YAMLFormat:
		str, err := template.YAML()
		if err != nil {
			logger.Log().Fatalf("Failed to marshal template - %v", err)
		}
		logger.Log().Println(str)
	case common.JSONFormat:
		str, err := template.JSON()
		if err != nil {
			logger.Log().Fatalf("Failed to marshal template - %v", err)
		}
		logger.Log().Println(str)
	}
}

func PrintTemplateList(format string, templates executable.TemplateList) {
	logger.Log().Debugf("listing %d templates", len(templates))
	switch common.NormalizeFormat(format) {
	case common.YAMLFormat:
		str, err := templates.YAML()
		if err != nil {
			logger.Log().Fatalf("Failed to marshal template list - %v", err)
		}
		logger.Log().Println(str)
	case common.JSONFormat:
		str, err := templates.JSON()
		if err != nil {
			logger.Log().Fatalf("Failed to marshal template list - %v", err)
		}
		logger.Log().Println(str)
	}
}
