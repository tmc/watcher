small filesystem poller and executor
====================================

example:

```shell
☭ ~ $ go get github.com/traviscline/watcher
☭ ~ $ watcher echo "file changed" &
[1] 35356
☭ ~ $ touch foo
☭ ~ $ running echo [file changed]
file changed

☭ ~ $ rm foo
☭ ~ $ running echo [file changed]
file changed
```
