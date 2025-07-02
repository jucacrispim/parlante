// go:build !test
package main

// notest

import (
	"flag"
	"fmt"
	"os"

	"github.com/jucacrispim/parlante"
	"github.com/jucacrispim/parlante/tui"
)

func main() {
	dbpath := flag.String("dbpath", parlante.DEFAULT_DB_PATH, "path for database file")
	flag.CommandLine.Parse(os.Args[1:])

	parlante.SetupDB(*dbpath)
	cs := parlante.ClientStorageSQLite{}
	ds := parlante.ClientDomainStorageSQLite{}
	cos := parlante.CommentStorageSQLite{}
	p := tui.NewTui(cs, ds, cos)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

}
