package validation

import "embed"

// schemas embeds the published JSON schema files for flow file types.
// These schemas are self-contained (all $ref resolved) and generated
// by tools/docsgen from the source YAML schemas in types/.
//
// Run "flow generate" (or "go run ./tools/docsgen") to refresh these files.
//
//go:embed *.json
var schemas embed.FS
