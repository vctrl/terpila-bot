package main

import "testing"

type test struct {
	input    int64
	expected string
}

func TestRaz(t *testing.T) {
	tests := []test{
		{1, "раз"},
		{2, "раза"},
		{3, "раза"},
		{4, "раза"},
		{5, "раз"},
		{6, "раз"},
		{7, "раз"},
		{8, "раз"},
		{9, "раз"},
		{10, "раз"},
		{11, "раз"},
		{12, "раз"},
		{13, "раз"},
		{14, "раз"},
		{23, "раза"},
		{784, "раза"},
		{8989, "раз"},
	}
	for _, tt := range tests {
		f := raz(tt.input)
		if f != tt.expected {
			t.Fatalf("test input %d: expected %s but was %s", tt.input, tt.expected, f)
		}
	}
}
