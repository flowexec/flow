---
title: Imported Executables Config Reference
---

# Imported Executables Config Reference

flow can automatically generate executables from scripts and Makefiles using special comments. 
Supported script types include shell scripts (`.sh`), batch files (`.bat`, `.cmd`), and PowerShell scripts (`.ps1`).
flow parses these comments during workspace synchronization and creates executable definitions that can be run 
like any other flow executable. See [Importing Executables](executables.md#importing-executables) for more details.

> [!NOTE] The configuration comments must be at the top of the script file or right above the Makefile target definition.
> - **Shell / PowerShell:** Use `# ` as the comment prefix (e.g., `# f:name=deploy`)
> - **Batch files:** Use `REM ` or `:: ` as the comment prefix (e.g., `REM f:name=deploy`)

## Supported Fields

| Field              | Description | Example |
|--------------------|-------------|---------|
| `name`             | Executable name | `f:name=deploy-app` |
| `verb`             | Action verb | `f:verb=deploy` |
| `description`      | Executable description | `f:description=Deploy to production` |
| `tag` or `tags`    | Pipe-separated tags | `f:tags=deployment\|production` |
| `alias`or `aliases` | Pipe-separated aliases | `f:aliases=prod-deploy\|deploy-prod` |
| `timeout`          | Execution timeout | `f:timeout=10m` |
| `visibility`       | Executable visibility | `f:visibility=private` |
| `dir`              | Working directory | `f:dir=//` |
| `logMode`          | Log output format | `f:logMode=json` |

### Environment Parameters

Define environment variables that will be available to your script with `f:params` or `f:param`:

::: code-group
```bash [Shell (.sh)]
#!/bin/bash
# f:name=deploy-with-secrets f:verb=deploy
# f:params=secretRef:api-key:API_TOKEN|prompt:Environment?:ENV_NAME|text:production:DEFAULT_ENV

echo "Deploying to $ENV_NAME with token: ${API_TOKEN:0:8}..."
```

```batch [Batch (.bat)]
@echo off
REM f:name=deploy-with-secrets f:verb=deploy
REM f:params=secretRef:api-key:API_TOKEN|prompt:Environment?:ENV_NAME|text:production:DEFAULT_ENV

echo Deploying to %ENV_NAME% with token: %API_TOKEN%...
```

```powershell [PowerShell (.ps1)]
# f:name=deploy-with-secrets f:verb=deploy
# f:params=secretRef:api-key:API_TOKEN|prompt:Environment?:ENV_NAME|text:production:DEFAULT_ENV

Write-Host "Deploying to $env:ENV_NAME with token: $($env:API_TOKEN.Substring(0,8))..."
```
:::

**Parameter Types:**
- `secretRef:secret-name:ENV_VAR` - Reference a vault secret
- `prompt:Question text:ENV_VAR` - Prompt user for input
- `text:static-value:ENV_VAR` - Set static value

### Command Line Arguments

Define command line arguments that users can pass when running the executable with `f:args` or `f:arg`:

::: code-group
```bash [Shell (.sh)]
#!/bin/bash
# f:name=build-app f:verb=build
# f:args=flag:dry-run:DRY_RUN|pos:1:VERSION|flag:verbose:VERBOSE

if [ "$DRY_RUN" = "true" ]; then
    echo "DRY RUN: Would build version $VERSION"
else
    echo "Building version $VERSION"
fi
```

```batch [Batch (.bat)]
@echo off
REM f:name=build-app f:verb=build
REM f:args=flag:dry-run:DRY_RUN|pos:1:VERSION|flag:verbose:VERBOSE

if "%DRY_RUN%"=="true" (
    echo DRY RUN: Would build version %VERSION%
) else (
    echo Building version %VERSION%
)
```

```powershell [PowerShell (.ps1)]
# f:name=build-app f:verb=build
# f:args=flag:dry-run:DRY_RUN|pos:1:VERSION|flag:verbose:VERBOSE

if ($env:DRY_RUN -eq "true") {
    Write-Host "DRY RUN: Would build version $env:VERSION"
} else {
    Write-Host "Building version $env:VERSION"
}
```
:::

**Argument Types:**
- `flag:flag-name:ENV_VAR` - Named flag (`--flag-name`)
- `pos:1:ENV_VAR` - Positional argument (position 1, 2, etc.)

## Configuration Syntax

**Single Line Format**

Multiple configurations can be defined on a single line:

```bash
# f:name=my-task f:verb=run f:timeout=5m f:visibility=private
```

**Multi-Line Format**

Configurations can be split across multiple lines for readability:

```bash
# f:name=complex-task
# f:verb=deploy
# f:description="Complex deployment with multiple parameters"
# f:params=secretRef:api-key:API_TOKEN
# f:params=prompt:Target environment?:ENV_NAME
# f:args=flag:dry-run:DRY_RUN
# f:args=pos:1:VERSION
```

**Multi-Line Descriptions**

For longer descriptions, use the multi-line description syntax:

```bash
# f:name=complex-deploy f:verb=deploy
# <f|description>
# Deploy application to production environment
# 
# This executable handles the complete deployment process including:
# - Database migrations
# - Service deployment
# - Health checks
# <f|description>
```
