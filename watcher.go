// watches the current directory for changes and runs the specificed program on change

package main

import (
	"flag"
	"fmt"
	"github.com/howeyc/fsnotify"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var help = `watcher [command to execute]`

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [flags] [command to execute and args]\n", os.Args[0])
	flag.PrintDefaults()
}

var verbose = flag.Bool("v", false, "verbose")
var recurse = flag.Bool("r", true, "recurse")
var quiet = flag.Int("quiet", 800, "quiet period after command execution in milliseconds")

func main() {
	flag.Usage = usage
	flag.Parse()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}
	cmd, args := flag.Args()[0], flag.Args()[1:]

	fileEvents := make(chan *fsnotify.FileEvent, 100)

	// start watchAndExecute goroutine
	go watchAndExecute(fileEvents, cmd, args)

	// pipe all events to fileEvents (for buffering and draining)
	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				fileEvents <- ev
				// @todo handle created/renamed/deleted dirs
			case err := <-watcher.Error:
				log.Println("fsnotify error:", err)
			}
		}
	}()

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	if *recurse {
		err = watchDirAndChildren(watcher, cwd); 
	} else {
		err = watcher.Watch(cwd)
	}
	if err != nil {
		log.Fatal(err)
	}
	<-make(chan struct{})
	watcher.Close()
}

// Execute cmd with args when a file event occurs
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

// Add dir and children (recursively) to watcher
func watchDirAndChildren(watcher *fsnotify.Watcher, dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			if *verbose {
				fmt.Fprintln(os.Stderr, "Watching", path)
			}
			if err := watcher.Watch(path); err != nil {
                            return err
                        }
		}
                return nil
	})
}

// Drain events from channel until a particular time
func drainUntil(until <-chan time.Time, c chan *fsnotify.FileEvent) {
	for {
		select {
		case <-c:
		case <-until:
			return
		}
	}
}
