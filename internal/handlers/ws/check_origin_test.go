package ws

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckOrigin_FailClosed(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name           string
		allowedOrigins []string
		origin         string
		want           bool
	}{
		{"empty allowlist rejects matching origin", nil, "https://app.example.com", false},
		{"empty allowlist rejects empty origin", nil, "", false},
		{"explicit origin matches", []string{"https://app.example.com"}, "https://app.example.com", true},
		{"explicit origin rejects other", []string{"https://app.example.com"}, "https://evil.example.com", false},
		{"explicit origin rejects empty", []string{"https://app.example.com"}, "", false},
		{"wildcard allows any non-empty origin", []string{"*"}, "https://anything.example", true},
		{"wildcard still rejects empty origin", []string{"*"}, "", false},
		{"comma-separated list is parsed", []string{"https://a.example,https://b.example"}, "https://b.example", true},
		{"trims whitespace", []string{"  https://a.example , https://b.example "}, "https://a.example", true},
		{"case-sensitive match", []string{"https://app.example.com"}, "https://APP.example.com", false},
	}

	logger := slog.New(slog.NewTextHandler(io_discard{}, nil))

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := NewHandler(nil, nil, logger, tc.allowedOrigins...)
			req := httptest.NewRequest(http.MethodGet, "/ws", nil)
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			if got := h.checkOrigin(req); got != tc.want {
				t.Fatalf("checkOrigin(origin=%q, allow=%v) = %v, want %v",
					tc.origin, tc.allowedOrigins, got, tc.want)
			}
		})
	}
}

// io_discard is a tiny stand-in for io.Discard that avoids pulling in `io`.
type io_discard struct{}

func (io_discard) Write(p []byte) (int, error) { return len(p), nil }
