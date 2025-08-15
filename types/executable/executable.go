package executable

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/flowexec/tuikit/types"
	"gopkg.in/yaml.v3"

	"github.com/flowexec/flow/internal/errors"
	"github.com/flowexec/flow/internal/utils"
	"github.com/flowexec/flow/types/common"
)

//go:generate go run github.com/atombender/go-jsonschema@v0.16.0 -et --only-models -p executable -o executable.gen.go --capitalization URI --capitalization URL executable_schema.yaml

const (
	TmpDirLabel = "f:tmp"
)

type ExecutableList []*Executable

func (e Directory) ExpandDirectory(
	wsPath, execPath, processTmpDir string,
	env map[string]string,
) (dir string, isTmpDir bool, err error) {
	if e == TmpDirLabel {
		if processTmpDir != "" {
			return processTmpDir, true, nil
		}

		td, err := os.MkdirTemp("", "flow")
		if err != nil {
			return "", false, err
		}
		return td, true, nil
	}

	return utils.ExpandDirectory(string(e), wsPath, execPath, env), false, nil
}

type ExecutableEnvironment struct {
	Params ParameterList `json:"params" yaml:"params"`
	Args   ArgumentList  `json:"args"   yaml:"args"`
}

func (e *ExecExecutableType) SetLogFields(fields map[string]interface{}) {
	e.logFields = fields
}

func (e *ExecExecutableType) GetLogFields() map[string]interface{} {
	return e.logFields
}

type enrichedExecutableList struct {
	Executables []*enrichedExecutable `json:"executables" yaml:"executables"`
}

type enrichedExecutable struct {
	*Executable

	ID              string `json:"id"              yaml:"id"`
	Ref             string `json:"ref"             yaml:"ref"`
	Namespace       string `json:"namespace"       yaml:"namespace"`
	Workspace       string `json:"workspace"       yaml:"workspace"`
	Flowfile        string `json:"flowfile"        yaml:"flowfile"`
	FullDescription string `json:"fullDescription" yaml:"fullDescription"`
}

func (e *Executable) SetContext(workspaceName, workspacePath, namespace, flowFilePath string) {
	e.workspace = workspaceName
	e.workspacePath = workspacePath
	e.namespace = namespace
	e.flowFilePath = flowFilePath
}

func (e *Executable) SetInheritedFields(flowFile *FlowFile) {
	e.MergeTags(flowFile.Tags)
	if e.Visibility == nil && flowFile.Visibility != nil {
		v := ExecutableVisibility(*flowFile.Visibility)
		e.Visibility = &v
	}
	var descFromFIle string
	if flowFile.DescriptionFile != "" {
		mdBytes, err := os.ReadFile(flowFile.DescriptionFile)
		if err != nil {
			descFromFIle += fmt.Sprintf("**error rendering description file**: %s", err)
		} else {
			descFromFIle += string(mdBytes)
		}
	}
	e.inheritedDescription = strings.Join([]string{flowFile.Description, descFromFIle}, "\n")
}

func (e *Executable) enriched() *enrichedExecutable {
	return &enrichedExecutable{
		Executable:      e,
		ID:              e.ID(),
		Ref:             e.Ref().String(),
		Namespace:       e.Namespace(),
		Workspace:       e.Workspace(),
		Flowfile:        e.FlowFilePath(),
		FullDescription: strings.TrimSpace(execDescriptionMarkdown(e, false)),
	}
}

func (e *Executable) YAML() (string, error) {
	yamlBytes, err := yaml.Marshal(e.enriched())
	if err != nil {
		return "", fmt.Errorf("failed to marshal executable - %w", err)
	}
	return string(yamlBytes), nil
}

func (e *Executable) JSON() (string, error) {
	jsonBytes, err := json.MarshalIndent(e.enriched(), "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal executable - %w", err)
	}
	return string(jsonBytes), nil
}

func (e *Executable) Markdown() string {
	return execMarkdown(e)
}

func (e *Executable) Ref() Ref {
	return Ref(fmt.Sprintf("%s %s", e.Verb, e.ID()))
}

func (e *Executable) ID() string {
	if e.Workspace() == "" {
		return "unk"
	}

	return NewExecutableID(e.Workspace(), e.Namespace(), e.Name)
}

func (e *Executable) Env() *ExecutableEnvironment {
	typeFields := []any{
		e.Exec,
		e.Launch,
		e.Request,
		e.Render,
		e.Serial,
		e.Parallel,
	}
	var execType any
	for _, field := range typeFields {
		if field == nil {
			continue
		}

		isPtr := reflect.ValueOf(field).Kind() == reflect.Ptr && !reflect.ValueOf(field).IsNil()
		isVal := reflect.ValueOf(field).Kind() != reflect.Ptr && !reflect.ValueOf(field).IsZero()
		if isPtr || isVal {
			execType = field
			break
		}
	}
	v := reflect.ValueOf(execType)
	if v.Kind() != reflect.Ptr {
		return nil
	}
	typeElem := v.Elem()
	execEnv := new(ExecutableEnvironment)
	for field := 0; field < typeElem.NumField(); field++ {
		if typeElem.Field(field).Kind() == reflect.Slice && !typeElem.Field(field).IsZero() {
			switch typeElem.Field(field).Interface().(type) {
			case ParameterList:
				execEnv.Params, _ = typeElem.Field(field).Interface().(ParameterList)
			case ArgumentList:
				execEnv.Args, _ = typeElem.Field(field).Interface().(ArgumentList)
			}
		}
	}
	return execEnv
}

func (e *Executable) AliasesIDs() []string {
	if len(e.Aliases) == 0 {
		return nil
	}

	if e.Workspace() == "" {
		return nil
	}
	aliases := make([]string, 0)
	for _, alias := range e.Aliases {
		aliases = append(aliases, NewExecutableID(e.Workspace(), e.Namespace(), alias))
	}
	return aliases
}

func (e *Executable) Workspace() string {
	return e.workspace
}

func (e *Executable) WorkspacePath() string {
	return e.workspacePath
}

func (e *Executable) Namespace() string {
	return e.namespace
}

func (e *Executable) FlowFilePath() string {
	return e.flowFilePath
}

const TimeoutOverrideEnv = "FLOW_DEFAULT_TIMEOUT"

func (e *Executable) SetDefaults() {
	if e.Verb == "" {
		e.Verb = "exec"
	}
	if e.Visibility == nil || *e.Visibility == "" {
		v := ExecutableVisibility(common.VisibilityPrivate)
		if e.Name == "" && e.Namespace() == "" {
			// Unnamed, workspace root executables are public by default
			v = ExecutableVisibility(common.VisibilityPublic)
		}
		e.Visibility = &v
	}

	if e.Timeout == nil {
		if v, ok := os.LookupEnv(TimeoutOverrideEnv); ok {
			if d, err := time.ParseDuration(v); err == nil {
				e.Timeout = &d
			}
		}
	}
}

func (e *Executable) Validate() error {
	if e == nil {
		return fmt.Errorf("executable undefined; try running `flow sync`")
	}

	if e.Verb == "" {
		return fmt.Errorf("verb cannot be empty")
	} else if err := e.Verb.Validate(); err != nil {
		return err
	}
	if strings.Contains(e.Name, " ") {
		return fmt.Errorf("name cannot contain spaces")
	}

	if e.Env() != nil {
		if err := e.Env().Args.Validate(); err != nil {
			return fmt.Errorf("args validation failed - %w", err)
		}
	}

	err := utils.ValidateOneOf(
		"executable type",
		e.Exec,
		e.Launch,
		e.Request,
		e.Render,
		e.Serial,
		e.Parallel,
	)
	if err != nil {
		return err
	}

	if e.Workspace() == "" {
		return fmt.Errorf("workspace was not set")
	}
	if e.flowFilePath == "" {
		return fmt.Errorf("flowFile path was not set")
	}

	return nil
}

func (e *Executable) NameEquals(name string) bool {
	return e.Name == name || slices.Contains(e.Aliases, name)
}

func (e *Executable) MergeTags(tags common.Tags) {
	e.Tags = slices.Compact(append(e.Tags, tags...))
}

// IsVisibleFromWorkspace returns true if the executable should be shown in terminal output for the given workspace.
func (e *Executable) IsVisibleFromWorkspace(workspaceFilter string) bool {
	matchesWsFiler := e.Workspace() == workspaceFilter || workspaceFilter == "" || workspaceFilter == WildcardWorkspace
	if e.Visibility == nil {
		return matchesWsFiler
	}
	switch common.Visibility(*e.Visibility) {
	case common.VisibilityPrivate:
		return matchesWsFiler
	case common.VisibilityPublic:
		return true
	case common.VisibilityInternal, common.VisibilityHidden:
		return false
	default:
		return false
	}
}

// IsExecutableFromWorkspace returns true if the executable can be executed from the given workspace.
func (e *Executable) IsExecutableFromWorkspace(workspaceFilter string) bool {
	matchesWsFiler := e.Workspace() == workspaceFilter || workspaceFilter == "" || workspaceFilter == WildcardWorkspace
	if e.Visibility == nil {
		return matchesWsFiler
	}
	switch common.Visibility(*e.Visibility) {
	case common.VisibilityPrivate, common.VisibilityInternal:
		return matchesWsFiler
	case common.VisibilityPublic:
		return true
	case common.VisibilityHidden:
		return false
	default:
		return false
	}
}

func (l ExecutableList) YAML() (string, error) {
	enriched := &enrichedExecutableList{}
	for _, exec := range l {
		enriched.Executables = append(enriched.Executables, exec.enriched())
	}
	yamlBytes, err := yaml.Marshal(enriched)
	if err != nil {
		return "", fmt.Errorf("failed to marshal executable list - %w", err)
	}
	return string(yamlBytes), nil
}

func (l ExecutableList) JSON() (string, error) {
	enriched := &enrichedExecutableList{}
	for _, exec := range l {
		enriched.Executables = append(enriched.Executables, exec.enriched())
	}
	jsonBytes, err := json.MarshalIndent(enriched, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal executable list - %w", err)
	}
	return string(jsonBytes), nil
}

func (l ExecutableList) Items() []*types.EntityInfo {
	items := make([]*types.EntityInfo, 0)
	for _, exec := range l {
		item := &types.EntityInfo{
			Header: exec.Ref().String(),
			Desc:   exec.Description,
			ID:     exec.Ref().String(),
		}
		if t := common.Tags(exec.Tags); len(t) > 0 {
			item.Desc = fmt.Sprintf("[%s]\n", t.PreviewString()) + exec.Description
		}
		items = append(items, item)
	}
	return items
}

func (l ExecutableList) Singular() string {
	return "executable"
}

func (l ExecutableList) Plural() string {
	return "executables"
}

func (l ExecutableList) FindByVerbAndID(verb Verb, id string) (*Executable, error) {
	_, _, name := MustParseExecutableID(id) // Assumes that ws and ns has already been filtered down
	filteredList := l.FilterByVerb(verb)
	var exec *Executable
	for _, e := range filteredList {
		if e.NameEquals(name) {
			exec = e
			break
		}
	}
	if exec != nil {
		return exec, nil
	}
	return nil, errors.ExecutableNotFoundError{Verb: string(verb), Name: name}
}

func (l ExecutableList) FilterByTags(tags common.Tags) ExecutableList {
	if len(tags) == 0 {
		return l
	}

	filteredExecs := make(ExecutableList, 0)
	for _, exec := range l {
		t := common.Tags(exec.Tags)
		if t.HasAnyTag(tags) {
			filteredExecs = append(filteredExecs, exec)
		}
	}
	return filteredExecs
}

func (l ExecutableList) FilterByVerb(verb Verb) ExecutableList {
	if verb == "" || verb == "*" {
		return l
	}

	if err := verb.Validate(); err != nil {
		return ExecutableList{}
	}

	filteredExecs := make(ExecutableList, 0)
	for _, exec := range l {
		if exec.Verb.Equals(verb) {
			filteredExecs = append(filteredExecs, exec)
		}
	}
	return filteredExecs
}

func (l ExecutableList) FilterBySubstring(str string) ExecutableList {
	if str == "" {
		return l
	}

	filteredExecs := make(ExecutableList, 0)
	for _, exec := range l {
		ref := exec.Ref().String()
		if strings.Contains(ref, str) || strings.Contains(exec.Description, str) {
			filteredExecs = append(filteredExecs, exec)
		} else { // search in aliases
			for _, alias := range exec.Aliases {
				if strings.Contains(alias, str) {
					filteredExecs = append(filteredExecs, exec)
					break
				}
			}
		}
	}
	return filteredExecs
}

func (l ExecutableList) FilterByWorkspace(ws string) ExecutableList {
	executables := make(ExecutableList, 0)
	for _, exec := range l {
		if exec.IsVisibleFromWorkspace(ws) {
			executables = append(executables, exec)
		}
	}

	if ws == "" || ws == WildcardWorkspace {
		return executables
	}

	filteredExecs := make(ExecutableList, 0)
	for _, exec := range executables {
		if exec.Workspace() == ws {
			filteredExecs = append(filteredExecs, exec)
		}
	}
	return filteredExecs
}

func (l ExecutableList) FilterByNamespace(ns string) ExecutableList {
	if ns == WildcardNamespace {
		return l
	}

	filteredExecs := make(ExecutableList, 0)
	for _, exec := range l {
		if exec.Namespace() == ns {
			filteredExecs = append(filteredExecs, exec)
		}
	}
	return filteredExecs
}

const (
	WildcardNamespace = "*"
	WildcardWorkspace = "*"
)

func MustParseExecutableID(id string) (workspace, namespace, name string) {
	if id == "" {
		return WildcardWorkspace, "", ""
	}

	parts := strings.Split(id, "/")
	switch len(parts) {
	case 1: // no workspace
		subparts := strings.Split(parts[0], ":")
		if len(subparts) == 1 { // no namespace
			return WildcardWorkspace, WildcardNamespace, subparts[0]
		} else if len(subparts) == 2 { // namespace AND name
			return WildcardWorkspace, subparts[0], subparts[1]
		}
	case 2: // workspace
		subparts := strings.Split(parts[1], ":")
		if len(subparts) == 1 { // no namespace
			return parts[0], WildcardNamespace, subparts[0]
		} else if len(subparts) == 2 {
			return parts[0], subparts[0], subparts[1]
		}
	}
	panic(fmt.Sprintf("invalid executable ID: %s", id))
}

func NewExecutableID(workspace, namespace, name string) string {
	var ws, ns string
	if namespace != "" && namespace != WildcardNamespace {
		ns = namespace
	}
	if workspace != "" && workspace != WildcardWorkspace {
		ws = workspace
	}

	switch {
	case ws == "":
		return "" // TODO: return error or log warning
	case ns == "" && name == "":
		return ws + "/" // nameless executable
	case ns != "" && name == "":
		return fmt.Sprintf("%s:", ns)
	case ns != "":
		return fmt.Sprintf("%s/%s:%s", ws, ns, name)
	default:
		return fmt.Sprintf("%s/%s", ws, name)
	}
}

// MarshalJSON is used to handle the custom marshaling of the Executable type.
// It enriches the Executable with additional fields for YAML and JSON output,
// and also handles the timeout field as a string for better readability in the output.
func (e *Executable) MarshalJSON() ([]byte, error) {
	output := make(map[string]interface{})

	type BaseExecutable Executable
	baseData, err := json.Marshal((*BaseExecutable)(e))
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(baseData, &output); err != nil {
		return nil, err
	}

	if e.Timeout != nil {
		output["timeout"] = e.Timeout.String()
	}

	// These are required in order to enrich the output with additional fields when
	// marshalling an ExecutableList
	output["id"] = e.ID()
	output["ref"] = e.Ref().String()
	output["namespace"] = e.Namespace()
	output["workspace"] = e.Workspace()
	output["flowfile"] = e.FlowFilePath()
	output["fullDescription"] = execDescriptionMarkdown(e, false)

	return json.Marshal(output)
}

// UnmarshalJSON is used to handle the custom unmarshaling of the Executable type.
// It parses the timeout field from a string to a time.Duration, allowing for
// more flexible input formats in JSON.
func (e *Executable) UnmarshalJSON(data []byte) error {
	type Alias Executable
	aux := &struct {
		*Alias
		Timeout string `json:"timeout,omitempty"`
	}{
		Alias: (*Alias)(e),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.Timeout != "" {
		duration, err := time.ParseDuration(aux.Timeout)
		if err != nil {
			return err
		}
		e.Timeout = &duration
	}
	return nil
}

func (r *RequestExecutableType) MarshalJSON() ([]byte, error) {
	type Alias RequestExecutableType
	aux := &struct {
		*Alias
		Timeout string `json:"timeout,omitempty"`
	}{
		Alias: (*Alias)(r),
	}
	if r.Timeout != 0 {
		aux.Timeout = r.Timeout.String()
	}
	return json.Marshal(aux)
}

func (r *RequestExecutableType) UnmarshalJSON(data []byte) error {
	type Alias RequestExecutableType
	aux := &struct {
		*Alias
		Timeout string `json:"timeout,omitempty"`
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.Timeout != "" {
		duration, err := time.ParseDuration(aux.Timeout)
		if err != nil {
			return err
		}
		r.Timeout = duration
	}
	return nil
}
