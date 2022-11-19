[![GoDoc](http://godoc.org/github.com/agschwender/autoreload?status.svg)](http://godoc.org/github.com/agschwender/autoreload)

# autoreload

`autoreload` provides a package and command for automatically reloading an executable when that executable changes. It intended to be used in a local development environment to reload the executable after it has been modified. An example use case would be to reload a go web app after you have edited the source code and recompiled the executable.

This approach can be useful when you prefer to manually rebuild your application instead of relying functionality that watches your source files, recompiles the application and finally restarts the it. In my experience, the manual approach to rebuilding allows greater control of when it is triggered and provides more visibility on when it completes so that you can know that your changes are present when you retest.

## Installation

`autoreload` can be used as a package that is integrated into your application or as a command that is supplied an executable to monitor.

### Installation via Package

To integrate the package into your application, follow the example below.

```
package main

import (
    "github.com/agschwender/autoreload"
)

func main() {
    // Application setup
    
    autoreload.New().Start()

    // Application run and waiting
}
```

See the [provided example](https://github.com/agschwender/autoreload/blob/main/example/main.go) for greater detail on how to integrate the package into your application.

### Installation via Command

To integrate the command into your application, you must first install the `autoreload` command:

```
$ go install github.com/agschwender/autoreload/autoreloader@v1.1.1
```

Once installed, you can then execute the command by supplying it with the executable you want it to monitor and restart. For example if you wanted to run your server command, it may look like this:

```
$ autoreloader server --port=8080
```

## Demo

You can verify the behavior of the package or command installation by using the provided `example` command.

In one terminal, compile the commands

```
$ go install ./...
```

In another terminal, run

```
$ example
2022/11/18 10:06:43 Starting application
2022/11/18 10:06:43 Auto-reload is enabled
2022/11/18 10:06:43 Starting HTTP server
```

Change the `example/main.go` file and then re-install, using the first terminal

```
$ go install ./...
```

You should see the reload happen in your second terminal

```
2022/11/18 10:06:57 Executable changed; reloading process
2022/11/18 10:06:57 Received change event, shutting down
2022/11/18 10:06:58 Starting application
2022/11/18 10:06:58 Auto-reload is enabled
2022/11/18 10:06:58 Starting HTTP server 2
```

Similarly, you can run via the `autoreloader` with the `example` commands build in reloading turned off.

In your second terminal, run

```
$ autoreloader example --autoreload=false
2022/11/18 10:10:11 Starting application
2022/11/18 10:10:11 Starting HTTP server
```

Again make a change to the `example/main.go` file and then re-install, using the first terminal

```
$ go install ./...
```

You should see the reload happen in your second terminal

```
2022/11/18 10:11:08 Executable changed; reloading process
2022/11/18 10:11:09 Killing process
2022/11/18 10:11:09 Starting application
2022/11/18 10:11:09 Starting HTTP server 2
```
