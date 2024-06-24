package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
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

// output sets the io.Writer for output
var output io.Writer = os.Stdout

var usage string = `
A url-shortening web server

This uses a simple csv file of short,long urls as a database.

Run with the -d/-development flag to run in development mode, providing
live template reloads. In development mode, the urls are also checked at
startup.
`

// getFlags parses flags
func getOptions() (Options, error) {
	var options Options
	var parser = flags.NewParser(&options, flags.Default)
	parser.Usage = usage

	if _, err := parser.Parse(); err != nil {
		if !flags.WroteHelp(err) {
			parser.WriteHelp(output)
		}
		return options, earlyExitError
	}
	if net.ParseIP(options.IPAddress) == nil {
		return options, errors.New("invalid ip address")
	}
	if _, err := strconv.Atoi(options.Port); err != nil {
		return options, errors.New("invalid network port")
	}
	if options.Timeout < (time.Second * 2) {
		return options, errors.New("timeout shorter than 2 seconds")
	}
	if options.Workers < 2 {
		return options, errors.New("at least one worker is needed")
	}
	return options, nil
}

func main() {
	options, err := getOptions()
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
	err = s.serve()
	if err != nil {
		fmt.Printf("server error %v", err)
		os.Exit(1)
	}
}
