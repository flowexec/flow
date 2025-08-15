## flow exec

Execute any executable by reference.

### Synopsis


Execute an executable where EXECUTABLE_ID is the target executable's ID in the form of 'ws/ns:name'.
The flow subcommand used should match the target executable's verb or one of its aliases.

If the target executable accept arguments, they can be passed in the form of flag or positional arguments.
Flag arguments are specified with the format 'flag=value' and positional arguments are specified as values without any prefix.


See https://flowexec.io/types/flowfile?id=executableverb for more information on executable verbs and https://flowexec.io/types/flowfile?id=executableref for more information on executable IDs.


#### Examples
**Execute a nameless flow in the current workspace with the 'install' verb**

flow install

**Execute a nameless flow in the 'ws' workspace with the 'test' verb**

flow test ws/

**Execute the 'build' flow in the current workspace and namespace**

flow exec build

flow run build  (Equivalent to the above since 'run' is an alias for the 'exec' verb)

**Execute the 'docs' flow with the 'show' verb in the current workspace and namespace**

flow show docs

**Execute the 'build' flow in the 'ws' workspace and 'ns' namespace**

flow exec ws/ns:build

**Execute the 'build' flow in the 'ws' workspace and 'ns' namespace with flag and positional arguments**

flow exec ws/ns:build flag1=value1 flag2=value2 value3 value4


```
flow exec EXECUTABLE_ID [args...] [flags]
```

### Options

```
  -h, --help                help for exec
  -m, --log-mode string     Log mode (text, logfmt, json, hidden)
  -p, --param stringArray   Set a parameter value by env key. (i.e. KEY=value) Use multiple times to set multiple parameters.This will override any existing parameter values defined for the executable.
```

### Options inherited from parent commands

```
  -L, --log-level string   Log verbosity level (debug, info, fatal) (default "info")
      --sync               Sync flow cache and workspaces
```

### SEE ALSO

* [flow](flow.md)	 - flow is a command line interface designed to make managing and running development workflows easier.

