<div align="center">
 <a href="https://github.com/arl/statsviz" title="Statsviz's Github repository.">
    <img src="https://raw.githubusercontent.com/arl/statsviz/readme-docs/logo.png?sanitize=true" width="100" height="auto"/>
 </a>
<br>
<br>
<br>

<p align="center">
  <a href="https://pkg.go.dev/github.com/arl/statsviz" title="Statsviz on pkg.go.dev">
    <img src="https://pkg.go.dev/badge/github.com/arl/statsviz" alt="Go Reference">
  </a>
  <a href="https://img.shields.io/github/tag/arl/statsviz.svg" title="Latest tag">
    <img src="https://img.shields.io/github/tag/arl/statsviz.svg" alt="Latest tag">
  </a>
  <a href="https://awesome.re/mentioned-badge.svg" title="Mentioned in Awesome Go">
    <img src="https://awesome.re/mentioned-badge.svg" alt="Mentioned in Awesome Go">
  </a>
</p>
<p align="center">
  <a href="https://github.com/arl/statsviz/actions">
    <img src="https://github.com/arl/statsviz/workflows/Tests-linux/badge.svg" alt="Linux CI">
  </a>
  <a href="https://github.com/arl/statsviz/actions">
    <img src="https://github.com/arl/statsviz/workflows/Tests-others/badge.svg" alt="Others CI">
  </a>
  <a href="https://codecov.io/gh/arl/statsviz">
    <img src="https://codecov.io/gh/arl/statsviz/branch/main/graph/badge.svg" alt="Codecov">
  </a>
</p>

</div>

# Statsviz


Visualize real time plots of your Go program runtime metrics, including heap, objects, goroutines, GC pauses, scheduler and more, in your browser.

<div align="center">
<img alt="statsviz ui" width="300px" height="auto" src="https://github.com/arl/statsviz/raw/readme-docs/window-light.png">
<img alt="statsviz ui" width="300px" height="auto" src="https://github.com/arl/statsviz/raw/readme-docs/window-dark.png">
</div>


- [Statsviz](#statsviz)
  - [Install](#install)
  - [Usage](#usage)
  - [Examples](#examples)
  - [How Does That Work?](#how-does-that-work)
  - [Documentation](#documentation)
    - [Go API](#go-api)
    - [Web User Interface](#web-user-interface)
    - [Plots](#plots)
    - [User Plots](#user-plots)
  - [Questions / Troubleshooting](#questions--troubleshooting)
  - [Contributing](#contributing)
  - [Changelog](#changelog)
  - [License: MIT](#license-mit)

## Install

Get the latest version:

```
go get github.com/arl/statsviz@latest
```


## Usage

Register `Statsviz` HTTP handlers with your application `http.ServeMux`.

```go
mux := http.NewServeMux()
statsviz.Register(mux)

go func() {
    log.Println(http.ListenAndServe("localhost:8080", mux))
}()
```

Open your browser at http://localhost:8080/debug/statsviz


## Examples

If you check any of the boxes below:
  - [ ] you use some HTTP framework
  - [ ] you want Statsviz to be located at `/my/path/to/statsviz` rather than `/debug/statsviz`
  - [ ] you want Statsviz under `https://` rather than `http://`
  - [ ] you want Statsviz behind some middleware

Then you should call `statsviz.NewServer()` (with or without options depending on your use case) in order to access the `Index()` and `Ws()` methods.

```go
srv, err := statsviz.NewServer(); // Create server or handle error
if err != nil { /* handle error */ }

// Do something with the handlers.
srv.Index()     // UI (dashboard) handler func
srv.Ws()        // Websocket handler func
```

Examples for the following cases, and more, are found in the [\_example](./_example/README.md) directory:

- use of `http.DefaultServeMux` or your own `http.ServeMux`
- wrap HTTP handler behind a middleware
- register the web page at `/foo/bar` instead of `/debug/statsviz`
- use `https://` rather than `http://`
- register Statsviz handlers with various Go HTTP libraries/frameworks:
  - [echo](https://github.com/labstack/echo/)
  - [fasthttp](https://github.com/valyala/fasthttp)
  - [fiber](https://github.com/gofiber/fiber/)
  - [gin](https://github.com/gin-gonic/gin)
  - and many others thanks to many contributors!


## How Does That Work?

Statsviz is made of two parts:

- The `Ws` serves a Websocket endpoint. When a client connects, your program's [runtime/metrics](https://pkg.go.dev/runtime/metrics) are sent to the browser, once per second, via the websocket connection.

- the `Index` http handler serves Statsviz user interface at `/debug/statsviz` at the address served by your program. When served, the UI connects to the Websocket endpoint and starts receiving data points.


## Documentation


### Go API

Check out the API reference on [pkg.go.dev](https://pkg.go.dev/github.com/arl/statsviz#section-documentation).

### Web User Interface


#### Top Bar

<img alt="webui-annotated" src="https://github.com/arl/statsviz/raw/readme-docs/webui-annotated.png">

##### Category Selector

<img alt="menu-categories" src="https://github.com/arl/statsviz/raw/readme-docs/menu-categories.png">

Each plot belongs to one or more categories. The category selector allows you to filter the visible plots by categories.

##### Visible Time Range

<img alt="menu-timerange" src="https://github.com/arl/statsviz/raw/readme-docs/menu-timerange.png">

Use the time range selector to define the visualized time span.

##### Show/Hide GC events

<img alt="menu-gc-events" src="https://github.com/arl/statsviz/raw/readme-docs/menu-gc-events.png">

Show or hide the vertical lines representing garbage collection events.

##### Pause updates

<img alt="menu-play" src="https://github.com/arl/statsviz/raw/readme-docs/menu-play.png">

Pause or resume the plot updates.


#### Plot Controls

<img alt="webui-annotated" src="https://github.com/arl/statsviz/raw/readme-docs/plot-controls-annotated.png">


### Plots

Which plots are visible depends on:
 - your Go version,since some plots are only available in newer versions.
 - what plot categories are currently selected. By default all plots are shown.

#### Allocation and Free Rate

<img width="50%" alt="alloc-free-rate" src="https://github.com/arl/statsviz/raw/readme-docs/plots/alloc-free-rate.png">

#### CGO Calls

<img width="50%" alt="cgo" src="https://github.com/arl/statsviz/raw/readme-docs/plots/cgo.png">

#### CPU (GC)

<img width="50%" alt="cpu-gc" src="https://github.com/arl/statsviz/raw/readme-docs/plots/cpu-gc.png">

#### CPU (Overall)

<img width="50%" alt="cpu-overall" src="https://github.com/arl/statsviz/raw/readme-docs/plots/cpu-overall.png">

#### CPU (Scavenger)

<img width="50%" alt="cpu-scavenger" src="https://github.com/arl/statsviz/raw/readme-docs/plots/cpu-scavenger.png">

#### Garbage Collection

<img width="50%" alt="garbage-collection" src="https://github.com/arl/statsviz/raw/readme-docs/plots/garbage collection.png">

#### GC Cycles

<img width="50%" alt="gc-cycles" src="https://github.com/arl/statsviz/raw/readme-docs/plots/gc-cycles.png">

#### GC Pauses

<img width="50%" alt="gc-pauses" src="https://github.com/arl/statsviz/raw/readme-docs/plots/gc-pauses.png">

#### GC Scan

<img width="50%" alt="gc-scan" src="https://github.com/arl/statsviz/raw/readme-docs/plots/gc-scan.png">

#### GC Stack Size

<img width="50%" alt="gc-stack-size" src="https://github.com/arl/statsviz/raw/readme-docs/plots/gc-stack-size.png">

#### Goroutines

<img width="50%" alt="goroutines" src="https://github.com/arl/statsviz/raw/readme-docs/plots/goroutines.png">

#### Heap (Details)

<img width="50%" alt="heap-details" src="https://github.com/arl/statsviz/raw/readme-docs/plots/heap (details).png">

#### Live Bytes

<img width="50%" alt="live-bytes" src="https://github.com/arl/statsviz/raw/readme-docs/plots/live-bytes.png">

#### Live Objects

<img width="50%" alt="live-objects" src="https://github.com/arl/statsviz/raw/readme-docs/plots/live-objects.png">

#### Memory Classes

<img width="50%" alt="memory-classes" src="https://github.com/arl/statsviz/raw/readme-docs/plots/memory-classes.png">

#### MSpan/MCache

<img width="50%" alt="mspan-mcache" src="https://github.com/arl/statsviz/raw/readme-docs/plots/mspan-mcache.png">

#### Mutex Wait

<img width="50%" alt="mutex-wait" src="https://github.com/arl/statsviz/raw/readme-docs/plots/mutex-wait.png">

#### Runnable Time

<img width="50%" alt="runnable-time" src="https://github.com/arl/statsviz/raw/readme-docs/plots/runnable-time.png">

#### Scheduling Events

<img width="50%" alt="sched-events" src="https://github.com/arl/statsviz/raw/readme-docs/plots/sched-events.png">

#### Size Classes

<img width="50%" alt="size-classes" src="https://github.com/arl/statsviz/raw/readme-docs/plots/size-classes.png">

#### GC Pauses

<img width="50%" alt="gc-pauses" src="https://github.com/arl/statsviz/raw/readme-docs/plots/gc-pauses.png">


### User Plots

Since `v0.6` you can add your own plots to Statsviz dashboard, in order to easily
visualize your application metrics next to runtime metrics.

Please see the [userplots example](_example/userplots/main.go).


## Questions / Troubleshooting

Either use GitHub's [discussions](https://github.com/arl/statsviz/discussions) or come to say hi and ask a live question on [#statsviz channel on Gopher's slack](https://gophers.slack.com/archives/C043DU4NZ9D).

## Contributing

Please use [issues](https://github.com/arl/statsviz/issues/new/choose) for bugs and feature requests.  
Pull-requests are always welcome!  
More details in [CONTRIBUTING.md](CONTRIBUTING.md).

## Changelog

See [CHANGELOG.md](./CHANGELOG.md).

## License: MIT

See [LICENSE](LICENSE)
