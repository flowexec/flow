package logs

import (
	"encoding/json"

	"github.com/flowexec/tuikit/io"
	"gopkg.in/yaml.v3"

	"github.com/flowexec/flow/internal/io/common"
	"github.com/flowexec/flow/internal/logger"
)

type entry struct {
	Args string `json:"args" yaml:"args"`
	Time string `json:"time" yaml:"time"`
	File string `json:"file" yaml:"file"`
}

type entryResponse struct {
	Logs []entry `json:"logs" yaml:"logs"`
}

func tuikitToEntry(e io.ArchiveEntry) entry {
	return entry{
		Args: e.Args,
		Time: e.Time.String(),
		File: e.Path,
	}
}

func marshalEntriesJSON(entries []io.ArchiveEntry) ([]byte, error) {
	entriesJSON := make([]entry, len(entries))
	for i, e := range entries {
		entriesJSON[i] = tuikitToEntry(e)
	}
	entriesResponse := entryResponse{Logs: entriesJSON}
	return json.MarshalIndent(entriesResponse, "", "  ")
}

func marshalEntriesYAML(entries []io.ArchiveEntry) ([]byte, error) {
	entriesYAML := make([]entry, len(entries))
	for i, e := range entries {
		entriesYAML[i] = tuikitToEntry(e)
	}
	entriesResponse := entryResponse{Logs: entriesYAML}
	return yaml.Marshal(entriesResponse)
}

func PrintEntries(format string, entries []io.ArchiveEntry) {
	logger.Log().Debugf("listing %d log entries", len(entries))
	switch common.NormalizeFormat(format) {
	case common.YAMLFormat:
		str, err := marshalEntriesYAML(entries)
		if err != nil {
			logger.Log().Fatalf("Failed to marshal log entries - %v", err)
		}
		logger.Log().Println(string(str))
	case common.JSONFormat:
		str, err := marshalEntriesJSON(entries)
		if err != nil {
			logger.Log().Fatalf("Failed to marshal log entries - %v", err)
		}
		logger.Log().Println(string(str))
	}
}
