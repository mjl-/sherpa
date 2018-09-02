# Sherpa

Sherpa is a Go library for creating a [sherpa API](https://www.ueber.net/who/mjl/sherpa/).

This library makes it trivial to export Go functions as a  sherap API with an http.Handler.

Your API will automatically be documented: cmd/sherpadoc reads your Go source, and exports function and type comments as API documentation.

See the [documentation](https://godoc.org/bitbucket.org/mjl/sherpa).


## Examples

A public sherpa API: https://www.sherpadoc.org/#https://www.sherpadoc.org/example/

That web application is [sherpaweb](https://bitbucket.org/mjl/sherpaweb]. It shows documentation for any sherpa API but also includes an API called Example for demo purposes.

[Ding](https://github.com/irias/ding/) is a more elaborate web application built with this library.


# About

Written by Mechiel Lukkien, mechiel@ueber.net.
Bug fixes, patches, comments are welcome.
MIT-licensed, see LICENSE.
cmd/sherpadoc/gopath.go originates from the Go project, see LICENSE-go for its BSD-style license.


# todo

- handler: write tests
- sherpadoc: write tests
- client: write tests
- sherpadoc: find out which go constructs people want to use that aren't yet implemented by sherpadoc

- when reading types from other packages (imported packages), we only look at GOPATH. vendor and modules are not taking into account, but we should.
