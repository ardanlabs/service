Contributing
============

First of all, thank you for considering to contribute to Statsviz!

Pull-requests are welcome!


## Go library

Statsviz Go public API surface is relatively limited, by design, and it's highly
unlikely that that will change. However new options can be added to
`statsviz.Register` and `statsviz.NewServer` without breaking compatibility.

Big changes should be discussed on the issue tracker prior to start working on
the code.

If you've decided to contribute, thank you so much, please comment on the
existing issue or create one stating what you want to tackle.


## User interface (html/css/javascript)

The user interface aims to be simple, light and minimal.

To bootstrap the UI for development:
 - cd to `internal/static`
 - run `npm install`
 - run `npm dev` and leave it running
 - in another terminal, cd to an example, for example `_example/default`
 - run `go mod edit -replace=github.com/arl/statsviz=../../` to build the
   example with your local version of the Go code. If you haven't touched to the
   Go code you can skip this step.

To build the production UI:
 - cd to `internal/static`
 - run `npm run build`
 - run `./scripts.zip.sh`
 - only commit `dist.zip`. `dist` directory is ignored.


Assets are located in the `internal/static` directory and are embedded with
[`go:embed`](https://pkg.go.dev/embed). To reduce the space taken by the assets
in the final binary, the `dist` directory is zipped into `dist.zip`. Use
`scripts/zip.sh` to do it. At runtime, when Statsviz serves the UI, the
`dist.zip` is then decompressed into a `fs.FS`, served via
`http.FileServerFS()`.


## `STATSVIZ_DEBUG`

Declare `STATSVIZ_DEBUG=1` environment variable when you develop in order to:
 - print websocket errors on standard output.
 - bypasses CORS checks

Obviously, this is not recommended for production use!

## Documentation

No contribution is too small. Improvements to code, comments or README
are welcome!


## Examples

There are many Go libraries to handle HTTP requests, routing, etc..

Feel free to add an example to show how to register Statsviz with your favourite
library.

To do so, please add a directory under `./_example`. For instance, if you want to add an
example showing how to register Statsviz within library `foobar`:

 - create a directory `./_example/foobar/`
 - create a file `./_example/foobar/main.go`
 - call `go example.Work()` as the first line of your example (see other
   examples). This forces the garbage collector to _do something_ so that
   Statsviz interface won't remain static when an user runs your example.
 - the code should be `gofmt`ed
 - the example should compile and run
 - when ran, Statsviz interface should be accessible at http://localhost:8080/debug/statsviz


Thank you!
