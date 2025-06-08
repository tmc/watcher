/*
Command watcher is a file watcher that executes a command when files change.

It can be used to automatically run tests, build code, or any other command
when files in a directory change.

# Usage:

	watcher [flags] [command to execute and args]

		-c	clear terminal before each run
		-depth int
			recursion depth (default 1)
		-dir string
			directory root to use for watching (default ".")
		-ignore string
			comma-separated list of glob patterns to ignore
		-quiet duration
			quiet period after command execution (default 800ms)
		-v	verbose
		-wait duration
			time to wait between change detection and exec (default 10ms)

# Features

  - Watch directories recursively with configurable depth
  - Ignore files based on glob patterns (e.g., `-ignore "*.tmp,*.log"`)
  - Automatically detects and watches newly created directories
  - Graceful termination on SIGINT/SIGTERM
  - Configurable quiet period to avoid rapid execution
  - Configurable wait time between file change and command execution

# Common Use Cases

  - Run tests when source files change
  - Trigger build steps in development
  - Refresh browsers or servers
  - Auto-compile code during development

# Example Use:

In Shell A:

```console
$ go install github.com/tmc/watcher@latest
$ mkdir /tmp/foo; cd /tmp/foo
$ watcher -v echo "triggered"
running echo triggered
triggered
```

Now, In Shell B:

```console
$ touch /tmp/foo/oi
```

Every time /tmp/foo changes the echo will be re-executed.

License: ISC
*/
package main

import (
	"fmt"
	"go/doc"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

//go:generate go run github.com/tmc/misc/gocmddoc@latest -o README.md

func usage() {
	fset := token.NewFileSet()
	pkgs, _ := parser.ParseDir(fset, ".", nil, parser.ParseComments)
	for _, pkg := range pkgs {
		if pkg.Name == "main" {
			d := doc.New(pkg, "./", 0)
			lines := strings.Split(d.Doc, "\n")
			fmt.Fprintln(os.Stderr, lines[0]) // First line
			for i, line := range lines {
				if strings.Contains(line, "Usage:") {
					for j := i; j < len(lines); j++ {
						if j > i && strings.HasPrefix(lines[j], "# ") {
							return // Stop at next heading
						}
						if lines[j] == "# Usage:" {
							fmt.Fprintln(os.Stderr, "\nUsage:")
						} else {
							fmt.Fprintln(os.Stderr, lines[j])
						}
					}
					return
				}
			}
		}
	}
}
