package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
)

var (
	verbose    = flag.Bool("v", false, "verbose")
	depth      = flag.Int("depth", 1, "recursion depth")
	dir        = flag.String("dir", ".", "directory root to use for watching")
	quiet      = flag.Duration("quiet", 800*time.Millisecond, "quiet period after command execution")
	wait       = flag.Duration("wait", 10*time.Millisecond, "time to wait between change detection and exec")
	ignoreFlag = flag.String("ignore", "", "comma-separated list of glob patterns to ignore")
	clear      = flag.Bool("c", false, "clear terminal before each run")
)

func main() {
	flag.Usage = usage
	flag.Parse()

	// Setup signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	var cmd string
	var args []string

	// If single argument that looks like a shell command, use sh -c
	singleArg := flag.Args()[0]
	if len(flag.Args()) == 1 && (strings.Contains(singleArg, ";") || strings.Contains(singleArg, "&&") || strings.Contains(singleArg, "||") || strings.Contains(singleArg, "|") || strings.Contains(singleArg, "=") || strings.Contains(singleArg, "$")) {
		cmd = "sh"
		args = []string{"-c", singleArg}
	} else {
		cmd, args = flag.Args()[0], flag.Args()[1:]
	}

	// Create the fsnotify watcher
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Error creating watcher: %v", err)
	}
	defer fsWatcher.Close()

	// Create a buffered channel for file events
	fileEvents := make(chan fsnotify.Event, 100)
	wg := &sync.WaitGroup{}

	// Create ignore patterns from comma-separated list
	var ignorePatterns []string
	if *ignoreFlag != "" {
		ignorePatterns = strings.Split(*ignoreFlag, ",")
	}

	// Start piping events from fsnotify to our channel
	wg.Add(1)
	go pipeEvents(ctx, wg, fsWatcher, fileEvents, ignorePatterns)

	// Start the command execution goroutine
	wg.Add(1)
	go watchAndExecute(ctx, wg, fileEvents, cmd, args)

	// Resolve the directory to watch
	watchDir, err := filepath.Abs(*dir)
	if err != nil {
		log.Fatalf("Error resolving watch directory: %v", err)
	}

	// Start watching the directory and its children
	if err := watchDirAndChildren(ctx, fsWatcher, watchDir, *depth); err != nil {
		log.Fatalf("Error watching directory: %v", err)
	}

	// Trigger initial run after wait delay
	time.Sleep(*wait)
	select {
	case fileEvents <- fsnotify.Event{Name: "startup", Op: fsnotify.Write}:
	case <-ctx.Done():
		return
	}

	// Wait for context cancellation
	<-ctx.Done()
	wg.Wait()
	if *verbose {
		fmt.Fprintln(os.Stderr, "Watcher shutting down")
	}
}

// Execute cmd with args when a file event occurs
func watchAndExecute(ctx context.Context, wg *sync.WaitGroup, fileEvents chan fsnotify.Event, cmd string, args []string) {
	defer wg.Done()

	var currentProcess *exec.Cmd
	var processMu sync.Mutex

	// Handle context cancellation to kill running process
	go func() {
		<-ctx.Done()
		processMu.Lock()
		if currentProcess != nil && currentProcess.Process != nil {
			currentProcess.Process.Kill()
		}
		processMu.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-fileEvents:
			if !ok {
				return
			}

			// Wait a bit between detecting the change and executing the command
			time.Sleep(*wait)

			if *verbose {
				fmt.Fprintln(os.Stderr, "File changed:", ev.Name)
			}

			// Clear terminal if requested
			if *clear {
				fmt.Print("\033[H\033[2J")
			}

			// Kill any existing process
			processMu.Lock()
			if currentProcess != nil && currentProcess.Process != nil {
				currentProcess.Process.Kill()
				currentProcess.Wait() // Wait for cleanup
			}

			// Execute command
			c := exec.CommandContext(ctx, cmd, args...)
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			c.Stdin = os.Stdin
			currentProcess = c
			processMu.Unlock()

			fmt.Fprintln(os.Stderr, "running", cmd, strings.Join(args, " "))
			if err := c.Run(); err != nil {
				if ctx.Err() == nil { // Only print if not caused by context cancellation
					fmt.Fprintln(os.Stderr, "error running:", err)
				}
			}

			processMu.Lock()
			currentProcess = nil
			processMu.Unlock()
			if *verbose {
				fmt.Fprintln(os.Stderr, "done.")
			}

			// Drain events during quiet period
			drainFor(ctx, *quiet, fileEvents)
		}
	}
}

// Add dir and children (recursively) to watcher
func watchDirAndChildren(ctx context.Context, w *fsnotify.Watcher, path string, depth int) error {
	// Add the directory to the watcher
	if err := w.Add(path); err != nil {
		return fmt.Errorf("error watching %s: %w", path, err)
	}

	if *verbose {
		fmt.Fprintln(os.Stderr, "Watching", path)
	}

	// Calculate the base path separator count for relative depth calculation
	baseNumSeps := strings.Count(path, string(os.PathSeparator))

	// Walk through all subdirectories
	return filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only interested in directories
		if !info.IsDir() {
			return nil
		}

		// Check if we should skip this directory based on depth
		pathDepth := strings.Count(walkPath, string(os.PathSeparator)) - baseNumSeps
		if pathDepth > depth {
			return filepath.SkipDir
		}

		// Don't re-add the root path
		if walkPath == path {
			return nil
		}

		// Add this directory to the watcher
		if *verbose {
			fmt.Fprintln(os.Stderr, "Watching", walkPath)
		}
		if err := w.Add(walkPath); err != nil {
			return fmt.Errorf("error watching %s: %w", walkPath, err)
		}

		return nil
	})
}

// pipeEvents sends valid events to the output channel, filtering based on ignore patterns
func pipeEvents(ctx context.Context, wg *sync.WaitGroup, w *fsnotify.Watcher, events chan fsnotify.Event, ignorePatterns []string) {
	defer wg.Done()
	defer close(events)

	// Handle directory creation events
	watchNewDirs := func(event fsnotify.Event) {
		if event.Has(fsnotify.Create) {
			// Check if this is a new directory
			info, err := os.Stat(event.Name)
			if err == nil && info.IsDir() {
				// Get the base directory to calculate depth
				wd, err := os.Getwd()
				if err != nil {
					log.Println("Error getting working directory:", err)
					return
				}

				// Only watch if within depth limit
				baseDir := *dir
				if !filepath.IsAbs(baseDir) {
					baseDir = filepath.Join(wd, baseDir)
				}

				baseNumSeps := strings.Count(baseDir, string(os.PathSeparator))
				pathDepth := strings.Count(event.Name, string(os.PathSeparator)) - baseNumSeps

				if pathDepth <= *depth {
					if *verbose {
						fmt.Fprintln(os.Stderr, "New directory detected:", event.Name)
					}
					if err := w.Add(event.Name); err != nil {
						log.Println("Error watching new directory:", err)
					}
				}
			}
		}
	}

	// Main event processing loop
	for {
		select {
		case <-ctx.Done():
			return
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			log.Println("fsnotify error:", err)
		case ev, ok := <-w.Events:
			if !ok {
				return
			}

			// Check for directory creation
			watchNewDirs(ev)

			// Skip ignored patterns
			if shouldIgnore(ev.Name, ignorePatterns) {
				continue
			}

			// Send the event
			select {
			case events <- ev:
			case <-ctx.Done():
				return
			}
		}
	}
}

// shouldIgnore checks if a file path matches any of the ignore patterns
func shouldIgnore(path string, patterns []string) bool {
	if len(patterns) == 0 {
		return false
	}

	// Get relative path from working directory
	wd, err := os.Getwd()
	if err != nil {
		log.Println("Error getting working directory:", err)
		return false
	}

	relPath, err := filepath.Rel(wd, path)
	if err != nil {
		log.Println("Error calculating relative path:", err)
		return false
	}

	// Check each pattern
	for _, pattern := range patterns {
		matched, err := filepath.Match(pattern, relPath)
		if err != nil {
			log.Println("Error matching pattern:", err)
			continue
		}
		if matched {
			return true
		}
	}

	return false
}

// drainFor drains events from channel until a time period has elapsed
func drainFor(ctx context.Context, drainUntil time.Duration, c chan fsnotify.Event) {
	timeout := time.After(drainUntil)
	for {
		select {
		case <-ctx.Done():
			return
		case <-c:
			// Drain the event
		case <-timeout:
			return
		}
	}
}
