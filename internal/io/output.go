package io

import (
	"fmt"
	"strings"
)

func TypesDocsURL(docType, docId string) string {
	if docId != "" {
		docId = "?id=" + strings.ToLower(docId)
	}
	return fmt.Sprintf("https://flowexec.io/types/%s%s", docType, docId)
}
