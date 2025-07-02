// go:build !test

package main

// notest
import (
	"flag"
	"os"

	"github.com/jucacrispim/parlante"
)

func main() {
	dbpath := flag.String("dbpath", parlante.DEFAULT_DB_PATH, "path for database file")
	host := flag.String("host", "0.0.0.0", "host to listen.")
	port := flag.Int("port", 8080, "port to listen.")
	certfile := flag.String("certfile", "", "Path for the tls certificate file")
	keyfile := flag.String("keyfile", "", "Path for the tls key file")
	loglevel := flag.String("loglevel", "info", "log level for the server")
	flag.CommandLine.Parse(os.Args[1:])
	c := parlante.Config{
		Host:         *host,
		Port:         *port,
		CertFilePath: *certfile,
		KeyFilePath:  *keyfile,
		DBPath:       *dbpath,
		LogLevel:     *loglevel,
	}
	parlante.SetupDB(c.DBPath)
	s := parlante.NewServer(c)
	s.Run()
}
