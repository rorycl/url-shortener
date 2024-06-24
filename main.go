package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jessevdk/go-flags"
)

// Options are the command line options
type Options struct {
	IPAddress   string        `short:"i" long:"ipAddress" default:"0.0.0.0" description:"ipaddress"`
	Port        string        `short:"p" long:"port" default:"8000" description:"port"`
	Development bool          `short:"d" long:"development" description:"run in development mode"`
	Timeout     time.Duration `short:"t" long:"timeout" default:"5s" description:"http client timeout"`
	Workers     uint          `short:"w" long:"workers" default:"8" description:"http client workers"`
}

var earlyExitError error = errors.New("early exit error")

var usage string = `
A url-shortening web server

This uses a simple csv file of short,long urls as a database.

Run with the -d/-development flag to run in development mode, providing
live template reloads. In development mode, the urls are also checked at
startup.
`

// getFlags parses flags
func getFlags() (Options, error) {
	var options Options
	var parser = flags.NewParser(&options, flags.Default)
	parser.Usage = usage

	if _, err := parser.Parse(); err != nil {
		if !flags.WroteHelp(err) {
			parser.WriteHelp(os.Stdout)
		}
		return options, earlyExitError
	}
	return options, nil
}

func main() {
	options, err := getFlags()
	if err != nil {
		os.Exit(1)
	}
	s, err := newServer(
		options.Development,
		options.IPAddress,
		options.Port,
		options.Timeout,
		int(options.Workers),
	)

	if err != nil {
		fmt.Printf("server setup error %v", err)
		os.Exit(1)
	}
	s.serve()
}
