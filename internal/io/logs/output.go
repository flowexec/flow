package logs

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	tuikitIO "github.com/flowexec/tuikit/io"
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

	// Content and its companions are populated only when log content is requested.
	Content              string `json:"content,omitempty"              yaml:"content,omitempty"`
	ContentTruncated     bool   `json:"contentTruncated,omitempty"     yaml:"contentTruncated,omitempty"`
	ContentTotalLines    int    `json:"contentTotalLines,omitempty"    yaml:"contentTotalLines,omitempty"`
	ContentReturnedLines int    `json:"contentReturnedLines,omitempty" yaml:"contentReturnedLines,omitempty"`
}

type recordsResponse struct {
	History []recordOutput `json:"history" yaml:"history"`
}

func toRecordOutput(r UnifiedRecord, content ContentOptions, includeContent bool) recordOutput {
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
		if includeContent {
			applyContent(&out, *r.LogEntry, content)
		}
	}
	return out
}

// applyContent reads the archive entry's output, slices it per opts, and populates the
// content fields on out. Read/filter errors are non-fatal — the metadata still stands.
func applyContent(out *recordOutput, entry tuikitIO.ArchiveEntry, opts ContentOptions) {
	raw, err := entry.Read()
	if err != nil {
		return
	}
	res, err := ExtractContent(raw, opts)
	if err != nil {
		return
	}
	out.Content = res.Content
	out.ContentTruncated = res.Truncated
	out.ContentTotalLines = res.TotalLines
	out.ContentReturnedLines = res.ReturnedLines
}

// PrintRecords outputs unified records in the specified format (json, yaml, or plain text).
// When includeContent is set, each record's log output is read and sliced per content.
func PrintRecords(format string, records []UnifiedRecord, content ContentOptions, includeContent bool) {
	out := make([]recordOutput, len(records))
	for i, r := range records {
		out[i] = toRecordOutput(r, content, includeContent)
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
			"%s  %-40s  %6s  %-10s  %s",
			r.StartedAt.Format(time.RFC3339),
			r.Ref,
			r.Duration.Round(time.Millisecond),
			StatusText(r),
			OriginText(r),
		))
	}
}

// PrintLastRecord outputs metadata and log content for a single record. In text mode the
// full log is printed (sliced per content when any content option is set); in json/yaml the
// content fields are populated only when includeContent is set.
func PrintLastRecord(
	format string, record UnifiedRecord, stdout io.Writer, content ContentOptions, includeContent bool,
) {
	out := toRecordOutput(record, content, includeContent)

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
			raw, err := record.LogEntry.Read()
			switch {
			case err != nil:
				_, _ = fmt.Fprintf(stdout, "error reading log: %v\n", err)
			case raw != "":
				res, extractErr := ExtractContent(raw, content)
				if extractErr != nil {
					_, _ = fmt.Fprintf(stdout, "error filtering log: %v\n", extractErr)
					break
				}
				_, _ = fmt.Fprint(stdout, res.Content)
				if res.Truncated {
					_, _ = fmt.Fprintf(
						stdout,
						"\n... (showing %d of %d lines)\n",
						res.ReturnedLines, res.TotalLines,
					)
				}
			}
		}
	}
}
