package main

import (
	"log"
	"os"
	"os/exec"
	"sync"

	"github.com/agschwender/autoreload"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Must supply a command to autoreload")
	}

	// Verify that the supplied command exists and generate an exec
	// command.
	path, err := exec.LookPath(os.Args[1])
	if err != nil {
		log.Fatalf("Cannot find executable: %s", os.Args[1])
	}

	cmd := exec.Command(path, os.Args[2:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	var wg sync.WaitGroup

	// Start the autoreloader monitor.
	autoreload.New(
		autoreload.WithCommand(os.Args[1]),
		autoreload.WithOnReload(func() {
			// When the application needs to reload, we must kill the
			// spawned command.

			log.Printf("Killing process")
			if err := cmd.Process.Kill(); err != nil {
				log.Fatalf("Failed to kill process: %v", err)
			}

			// Add to the wait group so that the autoreloader executable
			// will not exit due to the command being killed.
			wg.Add(1)
		}),
	).Start()

	// Run the supplied command
	err = cmd.Run()

	// If there is anything in the wait group, it means that the
	// autoreloader package is restarting the executable. Wait here
	// until it has completed that process.
	wg.Wait()

	// Maintain the error exit code of the supplied command.
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			os.Exit(int(exiterr.ExitCode()))
		} else {
			os.Exit(1)
		}
	}
}
