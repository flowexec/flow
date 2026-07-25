package logs

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/flowexec/flow/v2/internal/io/common"
	"github.com/flowexec/flow/v2/pkg/logger"
)

type recordOutput struct {
	Ref        string `json:"ref"                  yaml:"ref"`
	StartedAt  string `json:"startedAt"            yaml:"startedAt"`
	Duration   string `json:"duration"             yaml:"duration"`
	Status     string `json:"status"               yaml:"status"`
	ExitCode   int    `json:"exitCode"             yaml:"exitCode"`
	Error      string `json:"error,omitempty"      yaml:"error,omitempty"`
	LogFile    string `json:"logFile,omitempty"    yaml:"logFile,omitempty"`
	Command    string `json:"command,omitempty"    yaml:"command,omitempty"`
	Label      string `json:"label,omitempty"      yaml:"label,omitempty"`
	Source     string `json:"source,omitempty"     yaml:"source,omitempty"`
	ClientName string `json:"clientName,omitempty" yaml:"clientName,omitempty"`
	SessionID  string `json:"sessionId,omitempty"  yaml:"sessionId,omitempty"`
}

type recordsResponse struct {
	History []recordOutput `json:"history" yaml:"history"`
}

func toRecordOutput(r UnifiedRecord) recordOutput {
	out := recordOutput{
		Ref:        r.Ref,
		StartedAt:  r.StartedAt.Format(time.RFC3339),
		Duration:   r.Duration.Round(time.Millisecond).String(),
		Status:     string(CanonicalStatus(r)),
		ExitCode:   r.ExitCode,
		Error:      r.Error,
		Command:    r.Command,
		Label:      r.Label,
		Source:     r.Source,
		ClientName: r.ClientName,
		SessionID:  r.SessionID,
	}
	if r.LogEntry != nil {
		out.LogFile = r.LogEntry.Path
	}
	return out
}

// PrintRecords outputs unified records in the specified format (json, yaml, or plain text).
func PrintRecords(format string, records []UnifiedRecord) {
	out := make([]recordOutput, len(records))
	for i, r := range records {
		out[i] = toRecordOutput(r)
	}

	switch common.NormalizeFormat(format) {
	case common.JSONFormat:
		data, err := json.MarshalIndent(recordsResponse{History: out}, "", "  ")
		if err != nil {
			logger.Log().Fatalf("Failed to marshal records - %v", err)
		}
		logger.Log().Println(string(data))
	case common.YAMLFormat:
		data, err := yaml.Marshal(recordsResponse{History: out})
		if err != nil {
			logger.Log().Fatalf("Failed to marshal records - %v", err)
		}
		logger.Log().Println(string(data))
	default:
		if len(records) == 0 {
			logger.Log().Println("No execution history found.")
			return
		}
		printRecordsText(records)
	}
}

func printRecordsText(records []UnifiedRecord) {
	for _, r := range records {
		logger.Log().Println(fmt.Sprintf(
			"%s  %-40s  %6s  %s",
			r.StartedAt.Format(time.RFC3339),
			r.Ref,
			r.Duration.Round(time.Millisecond),
			StatusText(r),
		))
	}
}

// PrintLastRecord outputs metadata and log content for a single record.
func PrintLastRecord(format string, record UnifiedRecord, stdout io.Writer) {
	out := toRecordOutput(record)

	switch common.NormalizeFormat(format) {
	case common.JSONFormat:
		data, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			logger.Log().Fatalf("Failed to marshal record - %v", err)
		}
		_, _ = fmt.Fprintln(stdout, string(data))
	case common.YAMLFormat:
		data, err := yaml.Marshal(out)
		if err != nil {
			logger.Log().Fatalf("Failed to marshal record - %v", err)
		}
		_, _ = fmt.Fprint(stdout, string(data))
	default:
		_, _ = fmt.Fprintf(stdout, "Executable: %s\n", record.Ref)
		if record.Label != "" {
			_, _ = fmt.Fprintf(stdout, "Label:      %s\n", record.Label)
		}
		if record.Command != "" {
			_, _ = fmt.Fprintf(stdout, "Command:    %s\n", record.Command)
		}
		_, _ = fmt.Fprintf(stdout, "Time:       %s\n", record.StartedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(stdout, "Duration:   %s\n", record.Duration.Round(time.Millisecond))
		_, _ = fmt.Fprintf(stdout, "Status:     %s\n", StatusText(record))
		if record.Source != "" {
			_, _ = fmt.Fprintf(stdout, "Source:     %s\n", record.Source)
		}
		if record.ClientName != "" {
			_, _ = fmt.Fprintf(stdout, "Client:     %s\n", record.ClientName)
		}
		if record.SessionID != "" {
			_, _ = fmt.Fprintf(stdout, "Session:    %s\n", record.SessionID)
		}
		if record.Error != "" {
			_, _ = fmt.Fprintf(stdout, "Error:      %s\n", record.Error)
		}
		_, _ = fmt.Fprintln(stdout)

		if record.LogEntry != nil {
			content, err := record.LogEntry.Read()
			if err != nil {
				_, _ = fmt.Fprintf(stdout, "error reading log: %v\n", err)
			} else if content != "" {
				_, _ = fmt.Fprint(stdout, content)
			}
		}
	}
}
