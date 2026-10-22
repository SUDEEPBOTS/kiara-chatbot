package moderation

import (
	"testing"
)

func TestNormalizationAndDetection(t *testing.T) {
	f := NewFilter("", "")
	f.AddWord("fuck")
	f.AddWord("bhenchod")
	f.AddWord("kamina")

	testCases := []struct {
		input    string
		expected bool
	}{
		{"f.u.c.k", true},
		{"fu.ck", true},
		{"f_u_c_k", true},
		{"f u c k", true},
		{"fuuuuuck", true},
		{"f@ck", true},
		{"b.h.e.n.c.h.o.d", true},
		{"bhench0d", true},
		{"bhenchoddd", true},
		{"k.a.m.i.n.a", true},
		{"hello friend how are you", false},
		{"welcome to the group", false},
	}

	for _, tc := range testCases {
		bad, _ := f.IsAbusive(tc.input)
		if bad != tc.expected {
			t.Errorf("Input '%s': expected %v, got %v", tc.input, tc.expected, bad)
		}
	}
}
