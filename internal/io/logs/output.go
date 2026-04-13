package logs

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/flowexec/flow/internal/io/common"
	"github.com/flowexec/flow/pkg/logger"
)

type recordOutput struct {
	Ref       string `json:"ref"               yaml:"ref"`
	StartedAt string `json:"startedAt"         yaml:"startedAt"`
	Duration  string `json:"duration"          yaml:"duration"`
	ExitCode  int    `json:"exitCode"          yaml:"exitCode"`
	Error     string `json:"error,omitempty"   yaml:"error,omitempty"`
	LogFile   string `json:"logFile,omitempty" yaml:"logFile,omitempty"`
}

type recordsResponse struct {
	History []recordOutput `json:"history" yaml:"history"`
}

func toRecordOutput(r UnifiedRecord) recordOutput {
	out := recordOutput{
		Ref:       r.Ref,
		StartedAt: r.StartedAt.Format(time.RFC3339),
		Duration:  r.Duration.Round(time.Millisecond).String(),
		ExitCode:  r.ExitCode,
		Error:     r.Error,
	}
	if r.LogEntry != nil {
		out.LogFile = r.LogEntry.Path
	}
	return out
}

// PrintRecords outputs unified records in the specified format (json, yaml, or plain text).
func PrintRecords(format string, records []UnifiedRecord) {
	if len(records) == 0 {
		logger.Log().Println("No execution history found.")
		return
	}

	switch common.NormalizeFormat(format) {
	case common.JSONFormat:
		out := make([]recordOutput, len(records))
		for i, r := range records {
			out[i] = toRecordOutput(r)
		}
		data, err := json.MarshalIndent(recordsResponse{History: out}, "", "  ")
		if err != nil {
			logger.Log().Fatalf("Failed to marshal records - %v", err)
		}
		logger.Log().Println(string(data))
	case common.YAMLFormat:
		out := make([]recordOutput, len(records))
		for i, r := range records {
			out[i] = toRecordOutput(r)
		}
		data, err := yaml.Marshal(recordsResponse{History: out})
		if err != nil {
			logger.Log().Fatalf("Failed to marshal records - %v", err)
		}
		logger.Log().Println(string(data))
	default:
		printRecordsText(records)
	}
}

func printRecordsText(records []UnifiedRecord) {
	for _, r := range records {
		status := "ok"
		if r.ExitCode != 0 {
			status = fmt.Sprintf("exit(%d)", r.ExitCode)
		}
		logger.Log().Println(fmt.Sprintf(
			"%s  %-40s  %6s  %s",
			r.StartedAt.Format(time.RFC3339),
			r.Ref,
			r.Duration.Round(time.Millisecond),
			status,
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
		status := "ok"
		if record.ExitCode != 0 {
			status = fmt.Sprintf("exit(%d)", record.ExitCode)
		}

		_, _ = fmt.Fprintf(stdout, "Executable: %s\n", record.Ref)
		_, _ = fmt.Fprintf(stdout, "Time:       %s\n", record.StartedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(stdout, "Duration:   %s\n", record.Duration.Round(time.Millisecond))
		_, _ = fmt.Fprintf(stdout, "Status:     %s\n", status)
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
