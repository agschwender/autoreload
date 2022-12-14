// Package autoreload restarts a process if its executable changes.
package autoreload

import (
	"context"
	"errors"
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
	// Info is intended to log an informational message
	Info(string)

	// Error is intended to log an error message
	Error(string, error)
}

type defaultLogger struct{}

func (l *defaultLogger) Info(msg string) {
	log.Println(msg)
}

func (l *defaultLogger) Error(msg string, err error) {
	log.Printf("%s: %v\n", msg, err)
}

type noopLogger struct{}

func (l *noopLogger) Info(msg string)             {} // nolint: unparam
func (l *noopLogger) Error(msg string, err error) {} // nolint: unparam

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
// default, it will log using the built-in log package. When a nil value
// is supplied for the logger, logging will be disabled.
func WithLogger(logger Logger) option {
	if logger == nil {
		logger = &noopLogger{}
	}
	return func(autoReloader *AutoReloader) {
		autoReloader.logger = logger
	}
}

// WithMaxAttempts defines how many times the AutoReloader should
// attempt to reload the application. By default, this is 10. If the
// supplied maxAttempts is less than 1, it will be treated as 1.
func WithMaxAttempts(maxAttempts int) option {
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	return func(autoReloader *AutoReloader) {
		autoReloader.maxAttempts = maxAttempts
	}
}

// WithOnReload defines a callback that is executed just prior to
// reloading the application. This is useful for gracefully shutting
// down your application.
func WithOnReload(onReload onReloadFunc) option {
	if onReload == nil {
		onReload = func() {}
	}
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
				for i := 0; i < ar.maxAttempts; i++ {
					sleep(250*time.Millisecond, watcher.Events)
					if i == 0 {
						ar.onReload()
					}
					tryExec(ar.logger, execPath, os.Args, os.Environ())
				}
				fatal(ar.logger, "Failed to reload process", errors.New("max attempts reached"))
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
		fatal(logger, msg, err)
	}
}

func mustLookPath(logger Logger, name string) string {
	path, err := exec.LookPath(name)
	if err != nil {
		fatal(logger, fmt.Sprintf("Cannot find executable: %s", name), err)
	}
	return path
}

// sleep pauses the current goroutine for at least duration d, swallowing
// all fsnotify events received in the interim.
func sleep(d time.Duration, events chan fsnotify.Event) {
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
		fatal(logger, fmt.Sprintf("syscall.Exec: %s", argv0), err)
	}
	os.Exit(0)
}

func fatal(logger Logger, msg string, err error) {
	logger.Error(msg, err)
	os.Exit(1)
}
