/*
Copyright The Helm Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package urlutil

import (
	"net/url"
	"testing"
)

func TestURLJoin(t *testing.T) {
	tests := []struct {
		name, url, expect string
		paths             []string
	}{
		{name: "URL, one path", url: "http://example.com", paths: []string{"hello"}, expect: "http://example.com/hello"},
		{name: "Long URL, one path", url: "http://example.com/but/first", paths: []string{"slurm"}, expect: "http://example.com/but/first/slurm"},
		{name: "URL, two paths", url: "http://example.com", paths: []string{"hello", "world"}, expect: "http://example.com/hello/world"},
		{name: "URL, no paths", url: "http://example.com", paths: []string{}, expect: "http://example.com"},
		{name: "basepath, two paths", url: "../example.com", paths: []string{"hello", "world"}, expect: "../example.com/hello/world"},
	}

	for _, tt := range tests {
		if got, err := URLJoin(tt.url, tt.paths...); err != nil {
			t.Errorf("%s: error %q", tt.name, err)
		} else if got != tt.expect {
			t.Errorf("%s: expected %q, got %q", tt.name, tt.expect, got)
		}
	}
}

func TestSchemeHostAndPortMatches(t *testing.T) {
	for _, tt := range []struct {
		a, b  string
		match bool
	}{
		{"http://example.com", "http://example.com", true},
		{"https://example.com", "https://example.com", true},
		{"http://example.com", "https://example.com", false},
		{"https://example.com", "http://example.com", false},
		{"http://example.com:80", "http://example.com:80", true},
		{"https://example.com:443", "https://example.com:443", true},
		{"http://example.com:1234", "http://example.com:5678", false},
		{"https://example.com:1234", "https://example.com:5678", false},
		// The following lines are subject of change, currently only there
		// to ensure that the existing logic works as expected and the
		// upcoming fix / improvement works as wanted
		{"http://example.com:80", "http://example.com", false},
		{"https://example.com:443", "https://example.com", false},
	} {
		u1, _ := url.Parse(tt.a)
		u2, _ := url.Parse(tt.b)
		if tt.match != SchemeHostAndPortMatches(u1, u2) {
			t.Errorf("Expected %q==%q to be %t", tt.a, tt.b, tt.match)
		}
	}
}

func TestEqual(t *testing.T) {
	for _, tt := range []struct {
		a, b  string
		match bool
	}{
		{"http://example.com", "http://example.com", true},
		{"http://example.com", "http://another.example.com", false},
		{"https://example.com", "https://example.com", true},
		{"http://example.com/", "http://example.com", true},
		{"https://example.com", "http://example.com", false},
		{"http://example.com/foo", "http://example.com/foo/", true},
		{"http://example.com/foo//", "http://example.com/foo/", true},
		{"http://example.com/./foo/", "http://example.com/foo/", true},
		{"http://example.com/bar/../foo/", "http://example.com/foo/", true},
		{"/foo", "/foo", true},
		{"/foo", "/foo/", true},
		{"/foo/.", "/foo/", true},
		{"%/1234", "%/1234", true},
		{"%/1234", "%/123", false},
		{"/1234", "%/1234", false},
	} {
		if tt.match != Equal(tt.a, tt.b) {
			t.Errorf("Expected %q==%q to be %t", tt.a, tt.b, tt.match)
		}
	}
}

func TestExtractHostname(t *testing.T) {
	tests := map[string]string{
		"http://example.com":      "example.com",
		"https://example.com/foo": "example.com",

		"https://example.com:31337/not/with/a/bang/but/a/whimper": "example.com",
	}
	for start, expect := range tests {
		if got, _ := ExtractHostname(start); got != expect {
			t.Errorf("Got %q, expected %q", got, expect)
		}
	}
}
