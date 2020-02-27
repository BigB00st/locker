package utils

import "testing"

func TestStringInSlice(t *testing.T) {
	cases := []struct {
		inStr  string
		inList []string
		want   bool
	}{
		{"Hello", []string{"dlrow ,olleH", "Hello, World", "TEST123"}, true},
		{"Hello עולם", []string{"This Contains Hello עולם...", "AAAA"}, true},
		{"A", []string{}, false},
		{"Hello", []string{"dlrow ,olleH", "Hlo, World", "TEST123"}, false},
	}
	for _, c := range cases {
		got := StringInSlice(c.inStr, c.inList)
		if got != c.want {
			t.Errorf("StringInSlice(%q, %q) == %v, want %v", c.inStr, c.inList, got, c.want)
		}
	}
}
