package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestURLs(t *testing.T) {

	tests := []struct {
		input string
		isErr bool
		count int
	}{
		{
			input: "abc,http://def",
			isErr: false,
			count: 1,
		},
		{
			input: " abc, http://def",
			isErr: false,
			count: 1,
		},
		{
			input: " abc|http://def",
			isErr: true, // record does not have two fields
			count: 0,
		},
		{
			input: "abc, http://def\n abc,http://deg",
			isErr: true, // duplicate abc
			count: 0,
		},
		{
			input: "abc#, http://def",
			isErr: true, // su has non alpha/- chars
			count: 0,
		},
		{
			input: "a bc, http://def",
			isErr: true, // su has space
			count: 0,
		},
		{
			input: "abc, def",
			isErr: true, // ru does not have http
			count: 0,
		},
		{
			input: "abc, https://def\nghi,https://xyz",
			isErr: false,
			count: 2,
		},
		{
			// trailing \n\n
			input: "abc, https://def\nghi,https://xyz\n\n",
			isErr: false,
			count: 2,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("subtest_%d", i), func(t *testing.T) {
			t.Parallel()
			s := strings.NewReader(tt.input)
			m, err := urls(s)
			if err == nil && tt.isErr {
				t.Errorf("expected err for %v", m)
			}
			if err != nil && !tt.isErr {
				t.Errorf("unexpected err %v", err)
			}
			if err == nil {
				if len(m) != tt.count {
					t.Errorf("m len %d expected %d", len(m), tt.count)
				}
			}
		})
	}
}

// TestURLData tests the stored csv file
func TestURLData(t *testing.T) {
	f, err := os.Open("data/short-urls.csv")
	if err != nil {
		t.Fatal(err)
	}
	m, err := urls(f)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range m {
		t.Logf("%-40s : %s\n", k, v)
	}
}
