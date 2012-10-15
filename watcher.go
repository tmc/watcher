// watches the current directory for changes and runs the specificed program on change

package main

import (
	"flag"
	"fmt"
	"github.com/howeyc/fsnotify"
	"log"
	"os"
	"os/exec"
	"time"
)

var help = `watcher [command to execute]`

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [flags] [command to execute and args]\n", os.Args[0])
	flag.PrintDefaults()
}

var verbose = flag.Bool("v", false, "verbose")
var after = flag.Int("after", 100, "execute command after [after] milliseconds")

func main() {
	flag.Usage = usage
	flag.Parse()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}
	cmd, args := flag.Args()[0], flag.Args()[1:]

	done := make(chan bool)
	var event <-chan time.Time
	go func() {
		for {
			c := exec.Command(cmd, args...)
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			select {
			case <-event:
				fmt.Println("running", cmd, args)
				if err := c.Run(); err != nil {
					log.Println(err)
				}
				event = nil
			case <-watcher.Event:
				if *verbose {
					fmt.Println("File changed:", ev)
				}
				if event == nil {
					event = time.After(200 * time.Millisecond)
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
