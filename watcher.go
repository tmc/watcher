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
	"strings"
	"time"
)

var help = `watcher [command to execute]`

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [flags] [command to execute and args]\n", os.Args[0])
	flag.PrintDefaults()
}

var verbose = flag.Bool("v", false, "verbose")
var kill = flag.Bool("k", false, "kill the previously running program before starting a new one")
var depth = flag.Int("d", 1, "recursion depth")
var quiet = flag.Int("quiet", 800, "quiet period after command execution in milliseconds")

func main() {
	flag.Usage = usage
	flag.Parse()

	watcher, err := newWatcher()
	if err != nil {
		log.Fatal(err)
	}
	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}
	cmd, args := flag.Args()[0], flag.Args()[1:]

	fileEvents := make(chan interface{}, 100)

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
	err = watcher.watchDirAndChildren(cwd, *depth)
	if err != nil {
		log.Fatal(err)
	}
	<-make(chan struct{})
	watcher.Close()
}

type watcher struct {
	*fsnotify.Watcher
}

func newWatcher() (watcher, error) {
	fsnw, err := fsnotify.NewWatcher()
	return watcher{fsnw}, err
}

// Execute cmd with args when a file event occurs
func watchAndExecute(fileEvents chan interface{}, cmd string, args []string) {
	for {
		// execute command
		c := exec.Command(cmd, args...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Stdin = os.Stdin

		fmt.Fprintln(os.Stderr, "running", cmd, args)
		if err := c.Start(); err != nil {
			fmt.Fprintln(os.Stderr, "error running:", err)
		}
		if *verbose {
			fmt.Fprintln(os.Stderr, "done.")
		}
		// drain until quiet period is over
		drainFor(*quiet, fileEvents)
		ev := <-fileEvents
		if *verbose {
			fmt.Fprintln(os.Stderr, "File changed:", ev)
		}

		if *kill {
			if *verbose {
				fmt.Fprintln(os.Stderr, "Attempting to kill previous command")
			}
			if err := c.Process.Kill(); err != nil {
				fmt.Fprintf(os.Stderr, "Error killing proc(%d): %s\n", c.Process.Pid, err)
			}
		}

		// Make sure resources are cleaned up
		c.Wait()
	}
}

// Add dir and children (recursively) to watcher
func (w watcher) watchDirAndChildren(path string, depth int) error {
	if err := w.Watch(path); err != nil {
		return err
	}
	baseNumSeps := strings.Count(path, string(os.PathSeparator))
	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			pathDepth := strings.Count(path, string(os.PathSeparator)) - baseNumSeps
			if pathDepth > depth {
				return filepath.SkipDir
			}
			if *verbose {
				fmt.Fprintln(os.Stderr, "Watching", path)
			}
			if err := w.Watch(path); err != nil {
				return err
			}
		}
		return nil
	})
}

// Drain events from channel until a particular time
func drainFor(drainTimeMs int, c chan interface{}) {
	for {
		select {
		case <-c:
		case <-time.After(time.Duration(drainTimeMs) * time.Millisecond):
			return
		}
	}
}
