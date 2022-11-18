[![GoDoc](http://godoc.org/github.com/agschwender/autoreload?status.svg)](http://godoc.org/github.com/agschwender/autoreload)

# autoreload

`autoreload` provides a package and a command for automatically reloading an executable when that executable changes. It intended to be used in a local development environment to reload the executable after it has been modified as part of the development process. An example use case, would be to reload a go web app after you have edit the source code and recompiled the executable.

## Installation

`autoreload` can be used as either a package that is integrated into your application or as an executable that you run your application with. To use the package, add the following to your application:

```
autoreload.New().Start()
```

See the [provided example](https://github.com/agschwender/autoreload/blob/main/example/main.go) for greater detail on how to integrate the package into your application.

To use the executable, you must first install it

```
go install github.com/agschwender/autoreload/autoreloader@v1.0.0
```

You can then execute it by running the following

```
autoreloader my-application -v -a --arg
```
