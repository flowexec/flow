---
title: Imported Executables Config Reference
---

# Imported Executables Config Reference

flow can automatically generate executables from shell scripts and Makefiles using special comments. 
flow parses these comments during workspace synchronization and creates executable definitions that can be run 
like any other flow executable. See [Importing Executables](../guide/executables.md#importing-executables) for more details.

> [!NOTE] The configuration comments must be at the top of the shell script or right above the Makefile target definition.

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

```bash
#!/bin/bash
# f:name=deploy-with-secrets f:verb=deploy
# f:params=secretRef:api-key:API_TOKEN|prompt:Environment?:ENV_NAME|text:production:DEFAULT_ENV

echo "Deploying to $ENV_NAME with token: ${API_TOKEN:0:8}..."
```

**Parameter Types:**
- `secretRef:secret-name:ENV_VAR` - Reference a vault secret
- `prompt:Question text:ENV_VAR` - Prompt user for input
- `text:static-value:ENV_VAR` - Set static value

### Command Line Arguments

Define command line arguments that users can pass when running the executable with `f:args` or `f:arg`:

```bash
#!/bin/bash
# f:name=build-app f:verb=build
# f:args=flag:dry-run:DRY_RUN|pos:1:VERSION|flag:verbose:VERBOSE

if [ "$DRY_RUN" = "true" ]; then
    echo "DRY RUN: Would build version $VERSION"
else
    echo "Building version $VERSION"
fi
```

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
