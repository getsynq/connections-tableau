package main

import "testing"

func Test_cleanupUrl(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{
			url:  "https://reports.foo.io/#/site/",
			want: "https://reports.foo.io/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			if got := cleanupUrl(tt.url); got != tt.want {
				t.Errorf("cleanupUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}
