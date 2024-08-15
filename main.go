package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/creack/pty"
	"github.com/fsnotify/fsnotify"
	"github.com/muesli/cancelreader"
	"golang.org/x/term"

	"github.com/jhowrez/tui-hotreload/pkg/options"
)

var (
	initialTermState  *term.State
	cmd               *exec.Cmd
	printDebugAndExit bool
)

func init() {
	var appCfgPath string
	flag.StringVar(&appCfgPath, "c", "./app.yaml", "app options config path. Default ./app.yaml")
	flag.BoolFunc("d", "print debug information and exit", func(flagValue string) error {
		if flagValue == "false" {
			return nil
		}
		printDebugAndExit = true
		return nil
	})
	flag.Parse()
	options.OptionsInit(&appCfgPath)
}

func RunCommand() *exec.Cmd {
	appOptions := options.GetOptions()

	err := exec.Command("/bin/bash", "-c", appOptions.Command.Build).Run()
	if err != nil {
		log.Printf("failed to build: %s", err)
		return nil
	}
	cmd = exec.Command("/bin/bash", "-c", appOptions.Command.Exec)

	// Set stdin in raw mode.
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	if initialTermState == nil {
		initialTermState = oldState
	}

	ptmx, err := pty.Start(cmd)
	if err != nil {
		log.Fatal(err)
	}

	// Handle pty size.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
				log.Printf("error resizing pty: %s", err)
				return
			}
		}
	}()
	ch <- syscall.SIGWINCH // Initial resize.

	cancelReader, err := cancelreader.NewReader(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		_, _ = io.Copy(ptmx, cancelReader)
	}()

	go func() {
		_, _ = io.Copy(os.Stdout, ptmx)
		_ = term.Restore(int(os.Stdin.Fd()), initialTermState)
		ptmx.Close()
		cancelReader.Cancel()
	}()

	return cmd
}

func main() {
	if options.GetOptions().Watch.Root != "" {
		os.Chdir(options.GetOptions().Watch.Root)
	}

	if printDebugAndExit {
		paths := getAllWatchedFolders()
		fmt.Printf("The following %d paths will be watched:\n", len(paths))
		for _, path := range paths {
			fmt.Printf("- %s\n", path)
		}

		os.Exit(0)
	}

	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	cmd = RunCommand()

	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if event.Has(fsnotify.Write) {
					// isReloadingCommand = true
					if cmd != nil {
						syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
						cmd.Process.Wait()
						time.Sleep(time.Millisecond * 250)
					}
					cmd = RunCommand()
					// isReloadingCommand = false
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	paths := getAllWatchedFolders()
	for _, path := range paths {
		err = watcher.Add(path)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Block main goroutine forever.
	<-make(chan struct{})
}

func getAllWatchedFolders() []string {
	paths := []string{}
	for _, pathPattern := range options.GetOptions().Watch.Folders {
		// Add path.
		matches, err := BetterGlob(pathPattern)
		if err != nil {
			log.Fatal(err)
		}
		paths = append(paths, matches...)
	}
	return paths
}
