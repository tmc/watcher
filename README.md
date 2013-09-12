small filesystem watcher and executor
=====================================

watcher is a simple utility that waits for filesystem activity to execute a command


Example:

In Shell A:
```sh
☭ ~ $ go get github.com/tmc/watcher
☭ ~ $ watcher -h
usage: watcher [flags] [command to execute and args]
  -d=1: recursion depth
  -quiet=800: quiet period after command execution in milliseconds
  -v=false: verbose
☭ ~ $ mkdir /tmp/foo; cd /tmp/foo
☭ /tmp/foo $ watcher echo "triggered"
running echo [triggered]
triggered
```

Now, In Shell B:
```sh
☭ ~ $ touch /tmp/foo/oi
```

Every time /tmp/foo changes the echo will be re-executed.

I use this to run tests, trigger build steps, refresh browsers, etc.

License: ISC
