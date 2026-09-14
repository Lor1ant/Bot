package app

import "testing"

func TestNormalizeBaseURL(t *testing.T) {
	cases := map[string]string{
		"https://app.example.com":                 "https://app.example.com",
		"https://app.example.com/":                "https://app.example.com",
		"app.example.com":                         "https://app.example.com",
		"https://app.example.com/miniapp/":        "https://app.example.com",
		"https://https://app.example.com/miniapp": "https://app.example.com",
		"  https://app.example.com/miniapp/  ":    "https://app.example.com",
		"":                                        "",
	}
	for in, want := range cases {
		if got := normalizeBaseURL(in); got != want {
			t.Errorf("normalizeBaseURL(%q) = %q, want %q", in, got, want)
		}
	}
}
