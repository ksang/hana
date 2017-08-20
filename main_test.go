package main

import (
	"testing"
)

func TestParseDataSource(t *testing.T) {
	cases := []struct {
		config   string
		expected DataSourceType
	}{
		{"datasource:\n  unknown", UNKNOWN},
		{"datasource:\n  asaka", ASAKA},
		{"datasource:\n  Asaka", ASAKA},
	}

	for idx, c := range cases {
		res := ParseDataSource(c.config)
		if res != c.expected {
			t.Errorf("Case #%d, actual: %v, expected: %v", idx+1, res, c.expected)
		}
	}
}
