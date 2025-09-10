#!/bin/bash
set -euo pipefail

SCHEMAS_DIR="../docs/schemas"
OUTPUT_DIR="src-tauri/src/types/generated"

mkdir -p "$OUTPUT_DIR"

generate_schema() {
    local schema_path="$1"
    local schema_name=$(basename "$schema_path" .json | sed 's/_schema//')
    local output_file="$OUTPUT_DIR/${schema_name}.rs"

    if [ ! -f "$schema_path" ]; then
        echo -e "Schema file not found: ${schema_path}"
        return 1
    fi

    if ! cargo typify --version &>/dev/null; then
        echo -e "Installing cargo-typify..."
        cargo install cargo-typify@0.4.3
    fi

    if cargo typify "$schema_path" --output "$output_file"; then
        echo -e "Generated rust types for ${schema_name}"
    else
        echo -e "Failed to generate rust types for ${schema_name}"
        return 1
    fi
}

failed=0
schema_names=()

for schema_path in "$SCHEMAS_DIR"/*.json; do
    if [ ! -f "$schema_path" ]; then
        echo "No JSON schemas found in $SCHEMAS_DIR"
        exit 1
    fi

    schema_name=$(basename "$schema_path" .json | sed 's/_schema//')
    schema_names+=("$schema_name")

    if ! generate_schema "$schema_path"; then
        failed=1
    fi
done

{
    echo "// Generated module exports"
    for name in "${schema_names[@]}"; do
        echo "pub mod ${name};"
    done
} > "$OUTPUT_DIR/mod.rs"

echo -e "Generated mod.rs for rust types"

if [ $failed -eq 1 ]; then
    echo -e "Some rust schema generations failed"
    exit 1
else
    echo -e "All rust schemas generated successfully"
    exit 0
fi