import type { Meta, StoryObj } from "@storybook/react";
import { CodeHighlighter } from "./CodeHighlighter";

const meta: Meta<typeof CodeHighlighter> = {
  title: "Components/CodeHighlighter",
  component: CodeHighlighter,
  parameters: {
    layout: "centered",
  },
  tags: ["autodocs"],
  argTypes: {
    copyButton: {
      control: "boolean",
    },
    theme: {
      control: "select",
      options: ["light", "dark", "auto", "system"],
    },
  },
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Basic: Story = {
  args: {
    children: `#!/bin/bash
echo "Hello, World!"
ls -la`,
    copyButton: true,
    theme: "light",
  },
};

export const ComplexScript: Story = {
  args: {
    children: `#!/bin/bash
# This is a complex bash script example
set -euo pipefail

# Configuration
CONFIG_FILE="/etc/app/config.json"
LOG_FILE="/var/log/app.log"
BACKUP_DIR="/backups"

# Colors for output
RED='\\033[0;31m'
GREEN='\\033[0;32m'
YELLOW='\\033[1;33m'
NC='\\033[0m' # No Color

# Function to log messages
log() {
    echo -e "\${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')]\${NC} $1" | tee -a "\$LOG_FILE"
}

# Function to handle errors
error() {
    echo -e "\${RED}[ERROR]\${NC} $1" >&2
    exit 1
}

# Check if running as root
if [[ \$EUID -eq 0 ]]; then
   error "This script should not be run as root"
fi

# Process files
for file in *.txt; do
    if [ -f "\$file" ]; then
        log "Processing: \$file"

        # Backup original
        cp "\$file" "\$BACKUP_DIR/\$(basename \$file).backup"

        # Process the file
        if grep -q "important" "\$file"; then
            log "Found important content in \$file"
            # Add your processing logic here
        fi
    fi
done

# Cleanup old backups (older than 30 days)
find "\$BACKUP_DIR" -name "*.backup" -mtime +30 -delete

log "Script completed successfully"`,
    copyButton: true,
    theme: "dark",
  },
};

export const SimpleCommands: Story = {
  args: {
    children: `# Simple commands
pwd
whoami
date
echo "Current directory: $(pwd)"
echo "User: $(whoami)"
echo "Date: $(date)"`,
    copyButton: true,
    theme: "light",
  },
};

export const WithoutCopyButton: Story = {
  args: {
    children: `echo "This code block doesn't have a copy button"
ls -la
cat /etc/hosts`,
    copyButton: false,
    theme: "dark",
  },
};

export const EverforestTheme: Story = {
  args: {
    children: `#!/bin/bash
# This will use the system theme setting
echo "Using system theme"
echo "Current time: $(date)"
echo "System info: $(uname -a)"`,
    copyButton: true,
    theme: "everforest",
  },
};
