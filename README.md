small filesystem watcher and executor
=====================================

example:

```shell
☭ ~ $ go get github.com/traviscline/watcher
☭ /tmp/foo $ watcher echo "triggered"
running echo [triggered]
triggered
^Z
[1]+  Stopped                 watcher echo "triggered"
☭ /tmp/foo $ bg
[1]+ watcher echo "triggered" &
☭ /tmp/foo $ touch foo
running echo [triggered]
triggered
☭ /tmp/foo $ rm foo
running echo [triggered]
triggered
```
