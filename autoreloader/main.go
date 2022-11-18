package main

import (
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/agschwender/autoreload"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Must supply a command to autoreload")
	}

	path, err := exec.LookPath(os.Args[1])
	if err != nil {
		log.Fatalf("Cannot find executable: %s", os.Args[1])
	}

	autoreload.New(
		autoreload.WithCommand(os.Args[1]),
	).Start()

	err = syscall.Exec(path, os.Args[1:], os.Environ())
	if err != nil {
		if errno, ok := err.(syscall.Errno); ok {
			os.Exit(int(errno))
		} else {
			os.Exit(1)
		}
	}
}
