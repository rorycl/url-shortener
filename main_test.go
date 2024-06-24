package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestGetOptions(t *testing.T) {

	// override main 'output' stdout redirector
	var buf bytes.Buffer
	output = &buf
	showBuf := false

	// quieten os.Stderr
	stderr := os.Stderr
	os.Stderr = os.NewFile(0, os.DevNull)

	defer func() {
		output = os.Stdout
		os.Stderr = stderr
	}()

	tests := []struct {
		argString   string
		ip          string
		port        string
		development bool
		timeout     time.Duration
		workers     uint
		ok          bool
	}{
		{ // 0
			argString:   "<prog> -h",
			development: false,
			ok:          false, // actually osexits
		},
		{ // 1
			argString:   "<prog> -d",
			development: true,
			ok:          true,
		},
		{ // 2
			argString:   "<prog> -d -p 2000",
			development: true,
			port:        "2000",
			ok:          true,
		},
		{ // 3
			argString:   "<prog> -d -i 127.0.0.2",
			development: true,
			ip:          "127.0.0.2",
			ok:          true,
		},
		{ // 4
			argString:   "<prog> -d -t 20s",
			development: true,
			timeout:     time.Second * 20,
			ok:          true,
		},
		{ // 5
			argString:   "<prog> -d -w 12",
			development: true,
			workers:     12,
			ok:          true,
		},
		{ // 6
			argString: "<prog> -i hi",
			ok:        false,
		},
		{ // 7
			argString: "<prog> -p hi",
			ok:        false,
		},
		{ // 8
			argString: "<prog> -t 1s",
			ok:        false,
		},
		{ // 9
			argString: "<prog> -w 0",
			ok:        false,
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			os.Args = strings.Fields(tt.argString)
			options, err := getOptions()

			if err != nil && tt.ok == true {
				t.Errorf("unexpected error %v", err)
				return
			}

			if tt.ip != "" && tt.ip != options.IPAddress {
				t.Errorf("ip %s expected %s", options.IPAddress, tt.ip)
			}
			if tt.port != "" && tt.port != options.Port {
				t.Errorf("port %s expected %s", options.Port, tt.port)
			}
			if tt.development != options.Development {
				t.Errorf("development %t unexpected (%t)", options.Development, tt.development)
			}
			if tt.timeout != 0 && tt.timeout != options.Timeout {
				t.Errorf("timeout %v expected %v", options.Timeout, tt.timeout)
			}
			if tt.workers != 0 && tt.workers != options.Workers {
				t.Errorf("workers %v expected %v", options.Workers, tt.workers)
			}
			if showBuf {
				fmt.Println(buf.String())
			}

		})
	}
}
