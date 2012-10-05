// watches the current directory for changes and runs the specificed program on change

package main

import (
	"fmt"
	"github.com/howeyc/fsnotify"
	"log"
	"os"
	"os/exec"
)

var help = `watcher [command to execute]`

func main() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, help)
		os.Exit(1)
	}
	cmd, args := os.Args[1], os.Args[2:]

	done := make(chan bool)
	go func() {
		for {
			c := exec.Command(cmd, args...)
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			select {
			case <-watcher.Event:
				fmt.Println("running", cmd, args)
				if err := c.Run(); err != nil {
					log.Println(err)
				}
			case err := <-watcher.Error:
				log.Println("fsnotify error:", err)
			}
		}
	}()

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	err = watcher.Watch(wd)
	if err != nil {
		log.Fatal(err)
	}
	<-done
	watcher.Close()
}
