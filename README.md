simple filesystem watcher and executor
======================================

watcher is a simple utility that waits for filesystem activity to execute a command


Example:

In Shell A:
```sh
$ go get github.com/tmc/watcher
$ watcher -h
usage: watcher [flags] [command to execute and args]
  -depth int
        recursion depth (default 1)
  -dir string
        directory root to use for watching (default ".")
  -ignore string
        comma-separated list of glob patterns to ignore
  -quiet duration
        quiet period after command execution (default 800ms)
  -v    verbose
  -wait duration
        time to wait between change detection and exec (default 10ms)
$ mkdir /tmp/foo; cd /tmp/foo
$ watcher echo "triggered"
running echo triggered
triggered
```

Now, In Shell B:
```sh
$ touch /tmp/foo/oi
```

Every time /tmp/foo changes the echo will be re-executed.

## Features

- Watch directories recursively with configurable depth
- Ignore files based on glob patterns (e.g., `-ignore "*.tmp,*.log"`)
- Automatically detects and watches newly created directories
- Graceful termination on SIGINT/SIGTERM
- Configurable quiet period to avoid rapid execution
- Configurable wait time between file change and command execution

## Common Use Cases

- Run tests when source files change
- Trigger build steps in development
- Refresh browsers or servers
- Auto-compile code during development

License: ISC