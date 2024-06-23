package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"regexp"
	"strings"
)

var shortURLValidRegex *regexp.Regexp = regexp.MustCompile("^[-A-Za-z0-9]+$")

// urls makes a map of short urls "su" to redirection urls "ru" from a
// csv file in su,ru format.
//
// su operations:
// * trimmed of spaces
// * trailing "/" character removed
// (it might be sensible in future to force to lowercase)
//
// su checks:
// * no duplicate su values
// * no spaces
// * only letters, numbers and "-" character
//
// ru operations:
// * trimmed of spaces
//
// ru checks:
// * starts with http
func urls(r io.Reader) (map[string]string, error) {
	m := map[string]string{}
	c := csv.NewReader(r)
	for {
		var su, ru string
		record, err := c.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return m, fmt.Errorf("csv reading error: %v", err)
		}
		if len(record) != 2 {
			return m, fmt.Errorf("csv record does not have 2 fields: %v", record)
		}

		su, ru = record[0], record[1]

		// su operations
		su = strings.TrimSpace(su)
		su = strings.TrimRight(su, "/")

		// su checks
		if _, exists := m[su]; exists {
			return m, fmt.Errorf("short url %s already exists: %v", su, record)
		}
		if strings.Contains(su, " ") {
			return m, fmt.Errorf("short url %s has a space: %v", su, record)
		}
		if !shortURLValidRegex.MatchString(su) {
			return m, fmt.Errorf("short url %s has invalid characters: %v", su, record)
		}

		// ru operations
		ru = strings.TrimSpace(ru)

		// ru checks
		if strings.Index(ru, "http") != 0 {
			return m, fmt.Errorf("target %s does not start with http: %v", ru, record)
		}
		m[su] = ru
	}
	return m, nil
}
