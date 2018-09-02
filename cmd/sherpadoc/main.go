/*
Sherpadoc parses Go code and outputs sherpa documentation in JSON.

This documentation is provided to the sherpa HTTP handler to serve
as documentation through the _docs function.

Example:

	sherpadoc Awesome >awesome.json

Sherpadoc parses Go code, finds a struct named "Awesome", and gathers
documentation:

Comments above the struct are used as section documentation.  Fields
in section structs must are treated as subsections, and can in turn
contain subsections. These subsections and their methods are also
exported and documented in the sherpa API. Add a struct tag "sherpa"
to override the name of the subsection, for example `sherpa:"Another
Awesome API"`.

Comments above method names are function documentation. A synopsis
is automatically generated.

Types used as parameters or return values are added to the section
documentation where they are used. The comments above the type are
used, as well as the comments for each field in a struct.  The
documented field names know about the "json" struct field tags.

More eloborate example:

	sherpadoc
		-title 'Awesome API by mjl' \
		-replace 'time.Time string,*time.Time string,example.com/some/pkg.SomeType [] string' \
		path/to/awesome/code Awesome \
		>awesome.json

Most common Go code patterns for API functions have been implemented
in sherpadoc, but you may run into missing support.
*/
package main

import "bitbucket.org/mjl/sherpa/sherpadoc"

func main() {
	sherpadoc.Main()
}
