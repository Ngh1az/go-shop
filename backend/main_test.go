package main

import "testing"

func TestSanitizeFilename(t *testing.T) {
	got := sanitizeFilename("My Photo (1).PNG")
	want := "my-photo--1-.png"
	if got != want {
		t.Errorf("sanitizeFilename() = %q, muốn %q", got, want)
	}
}

func TestSanitizeFilenameChanPathTraversal(t *testing.T) {
	got := sanitizeFilename("../../etc/passwd")
	want := "passwd"
	if got != want {
		t.Errorf("sanitizeFilename() = %q, muốn %q", got, want)
	}
}
