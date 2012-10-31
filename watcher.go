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
var quiet = flag.Int("quiet", 800, "quiet period after command execution in milliseconds")

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

	fileEvents := make(chan *fsnotify.FileEvent, 100)
	done := make(chan bool)

	// start watchAndExecute goroutine
	go watchAndExecute(fileEvents, cmd, args)

	// pipe all events to fileEvents (for buffering and draining)
	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				fileEvents <- ev
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

func watchAndExecute(fileEvents chan *fsnotify.FileEvent, cmd string, args []string) {
	for {
		// execute command
		c := exec.Command(cmd, args...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Stdin = os.Stdin

		fmt.Fprintln(os.Stderr, "running", cmd, args)
		if err := c.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "error running:", err)
		}
		if *verbose {
			fmt.Fprintln(os.Stderr, "done.")
		}
		// drain until quiet period is over
		drainUntil(time.After(time.Duration(*quiet)*time.Millisecond), fileEvents)
		ev := <-fileEvents
		if *verbose {
			fmt.Fprintln(os.Stderr, "File changed:", ev)
		}
	}
}

func drainUntil(until <-chan time.Time, c chan *fsnotify.FileEvent) {
	for {
		select {
		case <-c:
		case <-until:
			return
		}
	}
}
