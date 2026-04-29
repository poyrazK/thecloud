package docker

import (
	"testing"
)

func TestValidateSnapshotPath(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"absolute clean path", "/var/lib/thecloud/snapshots/foo.tar.gz", false},
		{"deep nested clean path", "/var/lib/thecloud/snapshots/sub/dir/foo.tar.gz", false},
		{"empty path", "", true},
		{"relative path", "snapshots/foo.tar.gz", true},
		{"dot path", "./foo.tar.gz", true},
		{"parent traversal", "/var/lib/../etc/passwd", true},
		{"obfuscated parent traversal", "/var/lib/.././../etc/passwd", true},
		{"trailing slash is non-canonical", "/var/lib/thecloud/", true},
		{"double slash is non-canonical", "/var//lib/thecloud", true},
		{"NUL byte rejected", "/var/lib/thecloud/foo\x00.tar.gz", true},
		{"embedded dotdot segment", "/var/lib/thecloud/../../etc/passwd", true},
		{"root path rejected", "/", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cleaned, err := validateSnapshotPath(tc.input)
			gotErr := err != nil
			if gotErr != tc.wantErr {
				t.Fatalf("validateSnapshotPath(%q) err=%v wantErr=%v (cleaned=%q)", tc.input, err, tc.wantErr, cleaned)
			}
			if !tc.wantErr && cleaned != tc.input {
				t.Fatalf("expected cleaned path to equal input for canonical %q, got %q", tc.input, cleaned)
			}
		})
	}
}
