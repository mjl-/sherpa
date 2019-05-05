/*
Sherpaclient calls Sherpa API functions and prints Sherpa API documentation from the command-line.

Example:

	# all documentation for the API
	sherpaclient -doc https://www.sherpadoc.org/example/

	# documentation for just one function
	sherpaclient -doc https://www.sherpadoc.org/example/ sum

	# call a function
	sherpaclient https://www.sherpadoc.org/example/ sum 1 1

The parameters to a function must be valid JSON. Don't forget to quote the double quotes of your JSON strings!

	Usage: sherpaclient [options] baseURL function [param ...]
	  -doc
		show documentation for all functions or single function if specified
	  -info
		show the API descriptor
*/
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mjl-/sherpa"
	"github.com/mjl-/sherpa/client"
	"github.com/mjl-/sherpadoc"
)

var (
	printDoc  = flag.Bool("doc", false, "show documentation for all functions or single function if specified")
	printInfo = flag.Bool("info", false, "show the API descriptor")
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("sherpaclient: ")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: sherpaclient [options] baseURL function [param ...]\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(2)
	}

	url := args[0]
	args = args[1:]

	if *printDoc {
		if len(args) > 1 {
			flag.Usage()
			os.Exit(2)
		}
		doc(url, args)
		return
	}
	if *printInfo {
		if len(args) != 0 {
			flag.Usage()
			os.Exit(2)
		}
		info(url)
		return
	}

	if len(args) < 1 {
		flag.Usage()
		os.Exit(2)
	}
	function := args[0]
	args = args[1:]
	params := make([]interface{}, len(args))
	for i, arg := range args {
		err := json.Unmarshal([]byte(arg), &params[i])
		if err != nil {
			log.Fatalf("error parsing parameter %v: %s\n", arg, err)
		}
	}

	c, err := client.New(url, []string{})
	if err != nil {
		log.Fatal(err)
	}
	var result interface{}
	err = c.Call(context.Background(), &result, function, params...)
	if err != nil {
		switch serr := err.(type) {
		case *sherpa.Error:
			if serr.Code != "" {
				log.Fatalf("error %v: %s\n", serr.Code, serr.Message)
			}
		}
		log.Fatalf("error: %s\n", err)
	}
	err = json.NewEncoder(os.Stdout).Encode(&result)
	if err != nil {
		log.Fatal(err)
	}
}

func info(url string) {
	c, err := client.New(url, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("ID: %s\n", c.JSON.ID)
	fmt.Printf("Title: %s\n", c.JSON.Title)
	fmt.Printf("Version: %s\n", c.JSON.Version)
	fmt.Printf("BaseURL: %s\n", c.BaseURL)
	fmt.Printf("SherpaVersion: %d\n", c.JSON.SherpaVersion)
	fmt.Printf("Functions:\n")
	for _, fn := range c.Functions {
		fmt.Printf("- %s\n", fn)
	}
}

func doc(url string, args []string) {
	c, err := client.New(url, nil)
	if err != nil {
		log.Fatal(err)
	}

	var doc sherpadoc.Section
	cerr := c.Call(context.Background(), &doc, "_docs")
	if cerr != nil {
		log.Fatalf("fetching documentation: %s\n", cerr)
	}

	if len(args) == 1 {
		printFunction(&doc, args[0])
	} else {
		printSection(&doc)
	}
}

func printFunction(doc *sherpadoc.Section, function string) {
	for _, fn := range doc.Functions {
		if fn.Name == function {
			fmt.Println(fn.Docs)
		}
	}
	for _, subSec := range doc.Sections {
		printFunction(subSec, function)
	}
}

func printSection(sec *sherpadoc.Section) {
	fmt.Printf("# %s\n\n%s\n\n", sec.Name, sec.Docs)
	for _, fn := range sec.Functions {
		fmt.Printf("# %s()\n%s\n\n", fn.Name, fn.Docs)
	}
	for _, subSec := range sec.Sections {
		printSection(subSec)
	}
	fmt.Println("")
}
