Sherpa is a library in Go for providing and consuming Sherpa API's.

The Sherpa specification can be found at:

	https://www.ueber.net/who/mjl/sherpa/

This library makes it trivial to export Go functions as a Sherpa API, including documentation.

Use the sherpaweb tool to read API documentation and call methods (for testing purposes).


# Documentation

https://godoc.org/bitbucket.org/mjl/sherpa/

# Compiling

	go get bitbucket.org/mjl/sherpa/

# About

Written by Mechiel Lukkien, mechiel@ueber.net. Bug fixes, patches, comments are welcome.
MIT-licensed, see LICENSE.


# todo

- switch to using errors as part of return values. way more natural then panic all over...

- handler: write tests
- handler: write documentation
- handler: run jshint on the js code in sherpajs.go

- client: write tests

- sherpadocs: write it, a tool that reads a go package (source code) and generates a JSON document for returning as the `_docs` function.
