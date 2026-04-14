package mcp

// Output types for MCP tool responses. These structs are used with WithOutputSchema[T]()
// to generate JSON schemas for structured tool output. They mirror the CLI JSON output
// shapes but are decoupled from internal types to keep the MCP schema stable.

// FlowInfoOutput is the output of the get_info tool.
//
// This is intentionally lightweight — the full concepts/file-types guides and JSON schemas
// are NOT embedded by default to avoid consuming large amounts of LLM context. Clients should
// fetch the docs index (llms.txt) and individual guide pages as needed.
type FlowInfoOutput struct {
	CurrentContext CurrentContext `json:"currentContext"`
	// Summary is a compact platform description suitable for context priming.
	Summary string `json:"summary"`
	// DocsURL is the root of the hosted documentation site.
	DocsURL string `json:"docsUrl"`
	// LLMsTxtURL points to the llms.txt index of docs pages (per https://llmstxt.org/).
	LLMsTxtURL string `json:"llmsTxtUrl"`
	// SchemaURLs maps each file type to its JSON schema URL.
	SchemaURLs SchemaURLs `json:"schemaUrls"`
	// GuideURLs maps key topic names to their documentation URL.
	GuideURLs map[string]string `json:"guideUrls"`
}

// SchemaURLs lists the URLs of the JSON schemas for flow file types.
type SchemaURLs struct {
	FlowFile  string `json:"flowFile"`
	Workspace string `json:"workspace"`
	Template  string `json:"template"`
	Config    string `json:"config"`
}

type CurrentContext struct {
	Workspace     string `json:"workspace"`
	Namespace     string `json:"namespace"`
	Vault         string `json:"vault"`
	WorkspaceMode string `json:"workspaceMode"`
	WorkspacePath string `json:"workspacePath"`
}

// WorkspaceOutput is the output of the get_workspace tool.
type WorkspaceOutput struct {
	Name            string              `json:"name"`
	Path            string              `json:"path"`
	DisplayName     string              `json:"displayName,omitempty"`
	Description     string              `json:"description,omitempty"`
	FullDescription string              `json:"fullDescription,omitempty"`
	DescriptionFile string              `json:"descriptionFile,omitempty"`
	Tags            []string            `json:"tags,omitempty"`
	EnvFiles        []string            `json:"envFiles,omitempty"`
	Executables     *ExecutableFilter   `json:"executables,omitempty"`
	VerbAliases     map[string][]string `json:"verbAliases,omitempty"`
}

type ExecutableFilter struct {
	Included []string `json:"included,omitempty"`
	Excluded []string `json:"excluded,omitempty"`
}

// WorkspaceListOutput is the output of the list_workspaces tool.
type WorkspaceListOutput struct {
	Workspaces []WorkspaceOutput `json:"workspaces"`
	NextCursor string            `json:"nextCursor,omitempty"`
	TotalCount int               `json:"totalCount"`
}

// ExecutableOutput is the output of the get_executable tool.
type ExecutableOutput struct {
	ID              string   `json:"id"`
	Ref             string   `json:"ref"`
	Name            string   `json:"name"`
	Namespace       string   `json:"namespace"`
	Workspace       string   `json:"workspace"`
	FlowFile        string   `json:"flowfile"`
	Description     string   `json:"description,omitempty"`
	FullDescription string   `json:"fullDescription,omitempty"`
	Verb            string   `json:"verb"`
	Visibility      string   `json:"visibility,omitempty"`
	Timeout         string   `json:"timeout,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	Aliases         []string `json:"aliases,omitempty"`
}

// ExecutableListOutput is the output of the list_executables tool.
type ExecutableListOutput struct {
	Executables []ExecutableOutput `json:"executables"`
	NextCursor  string             `json:"nextCursor,omitempty"`
	TotalCount  int                `json:"totalCount"`
}

// ExecutionOutput is the output of the execute tool.
type ExecutionOutput struct {
	Output string `json:"output"`
}

// LogEntry represents a single execution log record.
type LogEntry struct {
	Ref       string `json:"ref"`
	StartedAt string `json:"startedAt"`
	Duration  string `json:"duration"`
	ExitCode  int    `json:"exitCode"`
	Error     string `json:"error,omitempty"`
	LogFile   string `json:"logFile,omitempty"`
}

// LogListOutput is the output of the get_execution_logs tool.
type LogListOutput struct {
	History    []LogEntry `json:"history"`
	NextCursor string     `json:"nextCursor,omitempty"`
	TotalCount int        `json:"totalCount"`
}

// SyncOutput is the output of the sync_executables tool.
type SyncOutput struct {
	Output string `json:"output"`
}

// SwitchWorkspaceOutput is the output of the switch_workspace tool.
type SwitchWorkspaceOutput struct {
	Output string `json:"output"`
}

// WriteFlowFileOutput is the output of the write_flowfile tool.
type WriteFlowFileOutput struct {
	Path        string   `json:"path"`
	Executables []string `json:"executables"`
	Overwritten bool     `json:"overwritten"`
}

// WorkspaceConfigOutput is the output of the get_workspace_config tool.
type WorkspaceConfigOutput struct {
	WorkspaceOutput
}
