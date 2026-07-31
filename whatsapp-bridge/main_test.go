package main

import "testing"

func TestExtractDirectPathFromURLPreservesMediaQuery(t *testing.T) {
	rawURL := "https://mmg.whatsapp.net/v/t62.7118-24/file.enc?ccb=11-4&oh=token&oe=123"
	want := "/v/t62.7118-24/file.enc?ccb=11-4&oh=token&oe=123"

	if got := extractDirectPathFromURL(rawURL); got != want {
		t.Fatalf("extractDirectPathFromURL() = %q, want %q", got, want)
	}
}

func TestExtractDirectPathFromURLKeepsFallback(t *testing.T) {
	rawURL := "not a media URL"

	if got := extractDirectPathFromURL(rawURL); got != rawURL {
		t.Fatalf("extractDirectPathFromURL() = %q, want fallback %q", got, rawURL)
	}
}
