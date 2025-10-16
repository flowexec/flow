---
title: Templates & Workflow Generation
---

# Templates & Workflow Generation

Templates let you generate new workflows and project scaffolding with interactive forms. 
Perfect for creating consistent project structures, operation workflows, or any repeatable automation pattern.

## Quick Start

Let's create a simple web app template:

```shell
# Create a template file
touch webapp.flow.tmpl
```

```yaml
# webapp.flow.tmpl
form:
  - key: "name"
    prompt: "What's your app name?"
    required: true
  - key: "port"
    prompt: "Which port should it run on?"
    default: "3000"

template: |
  executables:
    - verb: start
      name: "{{ name }}"
      exec:
        cmd: "npm start -- --port {{ form["port"] }}"
    - verb: build
      name: "{{ name }}"
      exec:
        cmd: "npm run build"
```

Register and use it:

```shell
# Register the template
flow template add webapp ./webapp.flow.tmpl

# Generate from template
flow template generate my-app --template webapp
```

## Template Components

Templates have four main parts:

### 1. Forms - Collect User Input

Forms define interactive prompts shown during generation:

```yaml
form:
  - key: "namespace"
    prompt: "Which namespace?"
    default: "default"
  - key: "replicas"
    prompt: "How many replicas?"
    default: "3"
    validate: "^[0-9]+$"  # Numbers only
  - key: "deploy"
    prompt: "Deploy immediately?"
    type: "confirm"       # Yes/no question
  - key: "image"
    prompt: "Container image?"
    required: true        # Must provide value
```

**Form field types:**
- `text` - Single line input (default)
- `multiline` - Multi-line text
- `masked` - Hidden input for passwords
- `confirm` - Yes/no question

### 2. Templates - Generate Flow Files

The main template creates your flow file:

```yaml
template: |
  executables:
    - verb: deploy
      name: "{{ name }}"
      exec:
        params:
          - envKey: "REPLICAS"
            text: "{{ form["replicas"] }}"
        cmd: kubectl apply -f deployment.yaml
    
    - verb: scale
      name: "{{ name }}"
      exec:
        cmd: kubectl scale deployment {{ name }} --replicas={{ form["replicas"] }}
```

### 3. Artifacts - Copy Supporting Files

Copy and optionally template additional files:

```yaml
artifacts:
  # Copy static files
  - srcName: "docker-compose.yml"
    dstName: "docker-compose.yml"
  
  # Template files (process with form data)
  - srcName: "deployment.yaml.tmpl"
    dstName: "deployment.yaml"
    asTemplate: true
  
  # Conditional copying
  - srcName: "helm-values.yaml"
    if: form["type"] == "helm"
```

### 4. Hooks - Run Commands

Execute commands before/after generation:

```yaml
preRun:
  - cmd: mkdir -p config
  - ref: validate environment

postRun:
  - cmd: chmod +x scripts/*.sh
  - ref: "deploy {{ .name }}"
    if: form["deploy"]
```

## Real-World Example

Here's a complete Kubernetes deployment template:

```yaml
form:
  - key: "namespace"
    prompt: "Deployment namespace?"
    default: "default"
  - key: "image"
    prompt: "Container image?"
    required: true
  - key: "replicas"
    prompt: "Number of replicas?"
    default: "3"
    validate: "^[1-9][0-9]*$"
  - key: "expose"
    prompt: "Expose via LoadBalancer?"
    type: "confirm"

artifacts:
  - srcName: "k8s-deployment.yaml.tmpl"
    dstName: "deployment.yaml"
    asTemplate: true
  - srcName: "k8s-service.yaml.tmpl"
    dstName: "service.yaml"
    asTemplate: true
    if: form["expose"]

postRun:
  - cmd: echo "Generated Kubernetes manifests"
  - cmd: kubectl apply -f .
    if: form["deploy"]

template: |
  executables:
    - verb: deploy
      name: "{{ name }}"
      exec:
        cmd: kubectl apply -f deployment.yaml -f service.yaml
    
    - verb: scale
      name: "{{ name }}"
      exec:
        params:
          - prompt: "New replica count?"
            envKey: "REPLICAS"
        cmd: kubectl scale deployment {{ name }} --replicas=$REPLICAS
    
    - verb: logs
      name: "{{ name }}"
      exec:
        cmd: kubectl logs -l app={{ name }} -f
```

## Template Management

See the [template command reference](../cli/flow_template.md) for all detailed commands and options.

### Register Templates

```shell
# From file
flow template add webapp ./templates/webapp.flow.tmpl

# List registered templates
flow template list

# View template details
flow template get -t webapp
```

### Generate from Templates

```shell
# Using registered template
flow template generate my-app --template webapp

# Using file directly
flow template generate my-app --file ./webapp.flow.tmpl

# Specify workspace and output directory
flow template generate my-app \
  --template webapp \
  --workspace my-workspace \
  --output ./apps/my-app
```

## Template Language

flow uses [Expr](https://expr-lang.org) language for all template evaluation, but with Go template syntax. 
You write templates using familiar `{{ }}` syntax, but the expressions inside are evaluated using Expr.

**Available Variables:**

| Variable | Description | Example |
|----------|-------------|---------|
| `name` | Generated file name | `{{ name }}` |
| `workspace` | Target workspace | `{{ workspace }}` |
| `form` | Form input values | `{{ form["replicas"] }}` |
| `env` | Environment variables | `{{ env["USER"] }}` |
| `os` | Operating system | `{{ os }}` |
| `arch` | System architecture | `{{ arch }}` |
| `workspacePath` | Full path to workspace | `{{ workspacePath }}` |
| `flowFilePath` | Full path to target flow file | `{{ flowFilePath }}` |
| `templatePath` | Path to template file | `{{ templatePath }}` |
| `directory` | Target directory | `{{ directory }}` |

**Template Examples:**
```yaml
# Basic variable access
template: |
  executables:
    - name: "{{ name }}"
      exec:
        cmd: echo "Hello from {{ name }}"

# Form data access
    - verb: deploy
      exec:
        cmd: kubectl apply -f {{ form["manifest"] }}

# Conditionals
{{ if form["type"] == "web" }}
    - verb: start
      exec:
        cmd: npm start
{{ end }}

# String functions
{{ upper(name) }}
{{ form["image"] | replace(":", "-") }}
```

**`if` fields in artifacts/hooks:**
For `if` fields in artifacts, preRun, and postRun sections, use Expr directly (no `{{ }}` needed):

```yaml
artifacts:
  - srcName: "web.conf"
    if: form["type"] == "web"
  - srcName: "api.conf" 
    if: form["type"] == "api" and len(form["endpoints"]) > 0

postRun:
  - cmd: ./deploy.sh
    if: form["deploy"] and form["environment"] == "production"
```

See the [Expr language documentation](https://expr-lang.org/docs/language-definition) for more information on Expr syntax and functions.
