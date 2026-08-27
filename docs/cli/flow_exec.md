## flow exec

Execute any executable by reference.

### Synopsis

Execute an executable where EXECUTABLE_ID is the target executable's ID in the form of 'ws/ns:name'.
The flow subcommand used should match the target executable's verb or one of its aliases.

If the target executable accepts arguments, use '--' to separate flow flags from executable arguments.
Flag arguments use standard '--flag=value' or '--flag value' syntax. Boolean flags can omit the value (e.g. '--verbose' implies true).
Positional arguments are specified as values without any prefix.

See https://flowexec.io/types/flowfile#executableverb for more information on executable verbs.
See https://flowexec.io/types/flowfile#executableref for more information on executable IDs.

```
flow exec EXECUTABLE_ID [-- args...] [flags]
```

### Examples

```

  # Execute a nameless flow in the current workspace with the 'install' verb
  flow install

  # Execute a nameless flow in the 'ws' workspace with the 'test' verb
  flow test ws/

  # Execute the 'build' flow in the current workspace and namespace
  flow exec build
  flow run build   # 'run' is an alias for the 'exec' verb

  # Execute the 'docs' flow with the 'show' verb
  flow show docs

  # Execute in a specific workspace and namespace
  flow exec ws/ns:build

  # Pass flag and positional arguments to the executable
  flow exec ws/ns:build -- --flag1=value1 --flag2=value2 value3 value4

```

### Options

```
  -b, --background           Run the executable in the background and return a run ID immediately.
      --cmd flow logs        Run an ad-hoc shell command through flow instead of a named executable. The command runs with the current workspace's environment and is recorded in flow logs. Repeat --cmd to run multiple commands in one invocation (see --mode).
      --dir string           Working directory for an ad-hoc command (defaults to the current directory). Only valid with --cmd.
  -h, --help                 help for exec
      --interpreter string   The interpreter to run ad-hoc --cmd commands with: 'sh' (default) or 'python'. Applies to every --cmd in the invocation.
      --label string         A short, human-readable label for an ad-hoc command (used in history). Only valid with --cmd.
  -m, --log-mode string      Log mode (text, logfmt, json, hidden)
      --mode string          How to run multiple --cmd commands: 'serial' (default) or 'parallel'. (default "serial")
  -p, --param stringArray    Set a parameter value by env key. (i.e. KEY=value) Use multiple times to set multiple parameters. This will override any existing parameter values defined for the executable.
      --spec flow logs       Run a transient executable from an inline definition (any type: exec, serial, parallel, request, render, launch). Accepts inline YAML/JSON, '@path' to read a file, or '-' to read stdin. The executable is not saved to disk but is recorded in flow logs.
      --workspace string     Workspace whose environment the ad-hoc/transient run should use (only with --cmd or --spec). Defaults to the workspace containing the run directory, then the current workspace. Does not change the global current workspace.
```

### Options inherited from parent commands

```
  -L, --log-level string   Log verbosity level (debug, info, fatal) (default "info")
      --sync               Sync flow cache and workspaces
```

### SEE ALSO

* [flow](flow.md)	 - flow is a command line interface designed to make managing and running development workflows easier.

