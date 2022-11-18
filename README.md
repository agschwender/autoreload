[![GoDoc](http://godoc.org/github.com/agschwender/autoreload?status.svg)](http://godoc.org/github.com/agschwender/autoreload)

# autoreload

`autoreload` provides a package for automatically reloading an executable when that executable changes. It intended to be used in a local development environment to reload the executable after it has been modified as part of the development process. An example use case, would be to reload a go web app after you have edit the source code and recompiled the executable.

## Installation

`autoreload` can be used as a package that is integrated into your application:

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
