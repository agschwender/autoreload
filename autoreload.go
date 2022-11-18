// Package autoreload restarts a process if its executable changes.
package autoreload

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
)

const defaultMaxAttempts = 10

type onReloadFunc func()

// Logger defines an interface for logging info and fatal errors out of
// the autoreloader process.
type Logger interface {
	Info(string)
	Fatal(string, error)
}

type defaultLogger struct{}

func (l *defaultLogger) Info(msg string) {
	log.Println(msg)
}

func (l *defaultLogger) Fatal(msg string, err error) {
	log.Fatalf("%s: %v\n", msg, err)
}

// AutoReloader provides functionality for reloading an application.
type AutoReloader struct {
	cmd         string
	logger      Logger
	maxAttempts int
	onReload    onReloadFunc

	ctx    context.Context
	cancel context.CancelFunc
}

type option func(*AutoReloader)

// New creates a new AutoReloader with the supplied options.
func New(opts ...option) AutoReloader {
	ctx, cancel := context.WithCancel(context.TODO())
	autoReloader := &AutoReloader{
		logger:      &defaultLogger{},
		maxAttempts: defaultMaxAttempts,
		onReload:    func() {},
		ctx:         ctx,
		cancel:      cancel,
	}
	for _, opt := range opts {
		opt(autoReloader)
	}
	return *autoReloader
}

// WithCommand defines the command executable that AutoReloader should
// watch. By default, this will be the currently running command.
func WithCommand(cmd string) option {
	return func(autoReloader *AutoReloader) {
		autoReloader.cmd = cmd
	}
}

// WithLogger defines the logger that the AutoReloader will use. By
// default, it will log using the built-in log package.
func WithLogger(logger Logger) option {
	return func(autoReloader *AutoReloader) {
		autoReloader.logger = logger
	}
}

// WithMaxAttempts defines how many times the AutoReloader should
// attempt to reload the application. By default, this is 10.
func WithMaxAttempts(maxAttempts int) option {
	return func(autoReloader *AutoReloader) {
		autoReloader.maxAttempts = maxAttempts
	}
}

// WithOnReload defines a callback that is executed just prior to
// reloading the application. This is useful for gracefully shutting
// down your application.
func WithOnReload(onReload onReloadFunc) option {
	return func(autoReloader *AutoReloader) {
		autoReloader.onReload = onReload
	}
}

// Start launches a goroutine that periodically checks if the modified
// time of the command has changed. If so, the binary is re-executed
// with the same arguments. This is a developer convenience and not
// intended to be started in a production environment.
func (ar AutoReloader) Start() {
	cmd := ar.cmd
	if cmd == "" {
		cmd = os.Args[0]
	}

	watchPath := mustLookPath(ar.logger, cmd)
	execPath := mustLookPath(ar.logger, os.Args[0])

	watcher, err := fsnotify.NewWatcher()
	must(ar.logger, err, "Failed to create file watcher")
	must(ar.logger, watcher.Add(watchPath), "Failed to watch file")

	go func() {
		for {
			select {
			case <-watcher.Events:
				ar.logger.Info("Executable changed; reloading process")
				sleep(250*time.Millisecond, watcher.Events)
				ar.onReload()
				for i := 0; i < ar.maxAttempts; i++ {
					tryExec(ar.logger, execPath, os.Args, os.Environ())
					sleep(250*time.Millisecond, watcher.Events)
				}
				ar.logger.Fatal("Failed to reload process", nil)
			case err := <-watcher.Errors:
				must(ar.logger, err, "Error watching file")
			case <-ar.ctx.Done():
				return
			}
		}
	}()
}

// Stop will stop the autoreloader from watching the executable and reloading it.
func (ar AutoReloader) Stop() {
	ar.cancel()
}

func must(logger Logger, err error, msg string) {
	if err != nil {
		logger.Fatal(msg, err)
	}
}

func mustLookPath(logger Logger, name string) string {
	path, err := exec.LookPath(name)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Cannot find executable: %s", name), err)
	}
	return path
}

// sleep pauses the current goroutine for at least duration d, swallowing
// all fsnotify events received in the interim.
func sleep(d time.Duration, events chan fsnotify.Event) { // nolint: unparam
	timer := time.After(d)
	for {
		select {
		case <-events:
		case <-timer:
			return
		}
	}
}

func tryExec(logger Logger, argv0 string, argv []string, envv []string) {
	if err := syscall.Exec(argv0, argv, envv); err != nil {
		if errno, ok := err.(syscall.Errno); ok {
			if errno == syscall.ETXTBSY {
				return
			}
		}
		logger.Fatal(fmt.Sprintf("syscall.Exec: %s", argv0), err)
	}
	os.Exit(0)
}
