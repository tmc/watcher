# watcher

[![Go Reference](https://pkg.go.dev/badge/github.com/tmc/watcher.svg)](https://pkg.go.dev/github.com/tmc/watcher)

Command watcher is a file watcher that executes a command when files change.

It can be used to automatically run tests, build code, or any other command when files in a directory change.
## Installation

<details>
<summary><b>Prerequisites: Go Installation</b></summary>

You'll need Go 1.23 or later. [Install Go](https://go.dev/doc/install) if you haven't already.

<details>
<summary><b>Setting up your PATH</b></summary>

After installing Go, ensure that `$HOME/go/bin` is in your PATH:

<details>
<summary><b>For bash users</b></summary>

Add to `~/.bashrc` or `~/.bash_profile`:
```bash
export PATH="$PATH:$HOME/go/bin"
```

Then reload your configuration:
```bash
source ~/.bashrc
```

</details>

<details>
<summary><b>For zsh users</b></summary>

Add to `~/.zshrc`:
```bash
export PATH="$PATH:$HOME/go/bin"
```

Then reload your configuration:
```bash
source ~/.zshrc
```

</details>

</details>

</details>

### Install

```console
go install github.com/tmc/watcher@latest
```

### Run directly

```console
go run github.com/tmc/watcher@latest [arguments]
```

## Usage:

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

## Features

  - Watch directories recursively with configurable depth
  - Ignore files based on glob patterns (e.g., \`-ignore "\*.tmp,\*.log"\`)
  - Automatically detects and watches newly created directories
  - Graceful termination on SIGINT/SIGTERM
  - Configurable quiet period to avoid rapid execution
  - Configurable wait time between file change and command execution

## Common Use Cases

  - Run tests when source files change
  - Trigger build steps in development
  - Refresh browsers or servers
  - Auto-compile code during development

## Example Use:

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
