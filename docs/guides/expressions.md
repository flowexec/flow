---
title: Expression Language
---

# Expression Language

flow uses the [Expr language](https://expr-lang.org) for dynamic expressions and template logic. The same language appears in four places — learn it once and it works everywhere.

## Where Expressions Are Used

| Surface | Syntax | Context variables | Shell `$()` |
|---------|--------|-------------------|:-----------:|
| Step `if` conditions | Bare expression (no delimiters) | `os`, `arch`, `env`, `store`, `ctx` | ✓ |
| `transformResponse` | Bare expression (no delimiters) | `body`, `code`, `status`, `headers` | |
| Template files (`.flow.tmpl`) | <span v-pre>`{{ expression }}`</span> delimiters | `name`, `form`, `env`, `os`, `arch`, … | |
| Render templates (render `.md`) | <span v-pre>`{{ expression }}`</span> delimiters | `env`, `data` | |

For the variables available in each surface, see the context-specific docs:
- **Step conditions** — [Advanced Workflows: Conditional Execution](./advanced#conditional-execution)
- **transformResponse** — [Executables: request](./executables#request---http-requests)
- **Template files** — [Templates & Workflow Generation: Template Language](./templating#template-language)
- **Render templates** — [Executables: render](./executables#render---dynamic-documentation)

## It's Not jq, Bash, or Go Templates

Expr is a sandboxed, typed, Go-native expression language. The key things that differ from other tools writers typically know:

- Field access on maps and parsed JSON uses **bracket notation**: `env["KEY"]`, `fromJSON(body)["field"]` — not `.key` or `$key`
- There are no pipes (`|`) for function application — use `upper(name)`, not `name | upper`
- `if` conditions use `==`, `not`, `and`, `or` — not `eq`, `ne`, `!`
- Shell execution via `$("command")` is available **only in step `if` conditions** — not in `transformResponse` or template surfaces

## Core Syntax

### Operators

| Operator | Meaning | Example |
|----------|---------|---------|
| `==`, `!=` | Equality | `os == "darwin"` |
| `<`, `>`, `<=`, `>=` | Comparison | `code >= 400` |
| `and`, `&&` | Logical AND | `os == "linux" and arch == "amd64"` |
| `or`, `\|\|` | Logical OR | `env["CI"] == "true" or env["DEBUG"] == "1"` |
| `not`, `!` | Negation | `not has(env, "CI")` |
| `+` | String concatenation | `"Hello, " + name` |
| `matches` | Regex match | `env["VERSION"] matches "^v[0-9]+"` |
| `in` | Membership | `"admin" in roles` |
| `? :` | Ternary | `code == 200 ? "ok" : "error"` |

### Map and Field Access

All map access uses bracket notation:

```
env["MY_VAR"]
form["replicas"]
headers["Content-Type"][0]
fromJSON(body)["items"]
```

Dot notation (`.field`) is **only** valid inside <span v-pre>`{{ range }}`</span> or <span v-pre>`{{ with }}`</span> template blocks — it does not work in bare expressions or at the top level of a template.

### JSON Handling

`body` and any raw string value is always a `string`. Call `fromJSON()` before accessing fields:

```
# Extract a field from a JSON response body
fromJSON(body)["name"]

# Nested field access
fromJSON(body)["user"]["email"]

# Store parsed result in a let binding to avoid reparsing
let data = fromJSON(body); data["status"] + " — " + data["message"]
```

### Let Bindings

`let` assigns a local name for use later in the same expression:

```
let parsed = fromJSON(body); parsed["id"]

let ns = ctx.namespace; "Deploying to " + ns + " on " + os
```

### Nil-Safe Conditionals

Use the ternary operator to guard against nil values:

```
value != nil ? string(value) : "default"

has(form, "image") ? form["image"] : "nginx:latest"
```

### Array Operations

`#` refers to the current element inside `map()` and `filter()`:

```
# Extract a field from each object in an array
map(fromJSON(body)["items"], #["name"])

# Filter to only enabled items
filter(fromJSON(body)["items"], #["enabled"] == true)

# Combine: extract IDs from enabled items
map(filter(fromJSON(body)["items"], #["enabled"]), string(#["id"]))
```

Combine `join()` and `map()` to format arrays as readable output:

```
join(map(fromJSON(body)["items"], #["name"]), "\n")
join(map(fromJSON(body)["tags"], string(#)), ", ")
```

### String Operations

```
# Concatenation
"prefix: " + string(value)

# Case conversion
upper(name)
lower(env["ENVIRONMENT"])

# Trimming
trim("  hello  ")

# Splitting and joining
split(env["TAGS"], ",")
join(split(env["TAGS"], ","), " | ")
```

### Shell Execution

In step `if` conditions, `$("command")` runs a POSIX shell command and returns its trimmed stdout as a string. The command runs using the same environment variables that are in scope for the executable.

```
# Check a tool is present
$("which kubectl") != ""

# Compare git branch
$("git rev-parse --abbrev-ref HEAD") == "main"

# Use output in a condition
$("cat VERSION") matches "^2\\."

# Combine with other context
os == "linux" and $("systemctl is-active docker") == "active"
```

> [!NOTE]
> `$()` is only available in step `if` conditions (serial and parallel executables). It is **not** available in `transformResponse`, template files, or render templates. If a command exits with a non-zero status, the expression returns an error.

## Built-in Functions

These functions are available in all Expr surfaces within flow.

### String and Type Functions

| Function | Description | Example |
|----------|-------------|---------|
| `fromJSON(s)` | Parse JSON string into value | `fromJSON(body)["key"]` |
| `toJSON(v)` | Serialize value to JSON string | `toJSON(fromJSON(body)["data"])` |
| `string(v)` | Convert value to string | `string(code)` |
| `int(v)` | Convert value to integer | `int(form["port"])` |
| `upper(s)` | Uppercase string | `upper(name)` |
| `lower(s)` | Lowercase string | `lower(env["ENV"])` |
| `trim(s)` | Strip leading/trailing whitespace | `trim(body)` |
| `split(s, sep)` | Split string into array | `split(env["TAGS"], ",")` |
| `join(arr, sep)` | Join array into string | `join(tags, ", ")` |
| `map(arr, expr)` | Transform each element (`#` = current) | `map(items, #["id"])` |
| `filter(arr, expr)` | Filter elements (`#` = current) | `filter(items, #["active"])` |
| `len(v)` | Length of string, array, or map | `len(form["name"]) > 0` |
| `has(map, key)` | Check if map has key | `has(env, "CI")` |
| `keys(map)` | List of map keys | `join(keys(env), ", ")` |
| `contains(s, sub)` | String contains substring | `contains(body, "error")` |
| `hasPrefix(s, pre)` | String starts with prefix | `hasPrefix(env["PATH"], "/usr")` |
| `hasSuffix(s, suf)` | String ends with suffix | `hasSuffix(name, "-prod")` |
| `replace(s, old, new)` | Replace substring | `replace(name, "-", "_")` |
| `trimPrefix(s, pre)` | Remove prefix | `trimPrefix(name, "app-")` |
| `trimSuffix(s, suf)` | Remove suffix | `trimSuffix(name, "-v2")` |

### File Helper Functions

These are available in all surfaces.

| Function | Description | Example |
|----------|-------------|---------|
| `fileExists(path)` | True if path exists (file or dir) | `fileExists("config.yaml")` |
| `dirExists(path)` | True if path is an existing directory | `dirExists(".git")` |
| `isFile(path)` | True if path is a regular file | `isFile("Makefile")` |
| `isDir(path)` | True if path is a directory | `isDir("node_modules")` |
| `basename(path)` | Filename component of a path | `basename("/home/user/file.txt")` → `"file.txt"` |
| `dirname(path)` | Directory component of a path | `dirname("/home/user/file.txt")` → `"/home/user"` |
| `readFile(path)` | Read file contents as a string | `readFile(".version")` |
| `fileSize(path)` | File size in bytes | `fileSize("output.log") > 0` |
| `fileModTime(path)` | File modification time | `fileModTime("lock.pid")` |
| `fileAge(path)` | Duration since file was last modified | `fileAge("cache.json")` |

### Shell Execution Function

Only available in step `if` conditions.

| Function | Description | Example |
|----------|-------------|---------|
| `$("cmd")` | Run a shell command, return trimmed stdout | `$("git branch --show-current")` |

> [!NOTE]
> This table covers the most commonly used functions. See the [Expr Language Definition](https://expr-lang.org/docs/language-definition) for the complete built-in reference.

## Common Patterns

**Extract and transform a JSON field:**
```
upper(fromJSON(body)["status"])
```

**Format a list from an API response:**
```
join(map(fromJSON(body)["items"], #["name"]), "\n")
```

**Conditional value with fallback:**
```
has(form, "region") ? form["region"] : "us-east-1"
```

**Check environment before running:**
```
env["DEPLOY_ENV"] == "production" and has(env, "API_KEY")
```

**Multi-step expression with let:**
```
let items = fromJSON(body)["results"]; join(map(items, #["id"] + ": " + #["name"]), "\n")
```

**Run a shell command to gate a step (step `if` only):**
```
$("git rev-parse --abbrev-ref HEAD") == "main" and env["CI"] == "true"
```
