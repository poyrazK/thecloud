package routing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPatternMatcher(t *testing.T) {
	tests := []struct {
		pattern        string
		path           string
		shouldMatch    bool
		expectedParams map[string]string
	}{
		{
			pattern:     "/users/{id}",
			path:        "/users/123",
			shouldMatch: true,
			expectedParams: map[string]string{
				"id": "123",
			},
		},
		{
			pattern:     "/users/{id}",
			path:        "/users/123/posts",
			shouldMatch: false,
		},
		{
			pattern:     "/users/{id}/posts/{pid}",
			path:        "/users/123/posts/456",
			shouldMatch: true,
			expectedParams: map[string]string{
				"id":  "123",
				"pid": "456",
			},
		},
		{
			pattern:     "/api/v{version}/*",
			path:        "/api/v1/posts",
			shouldMatch: true,
			expectedParams: map[string]string{
				"version": "1",
			},
		},
		{
			pattern:        "/files/*",
			path:           "/files/images/photo.jpg",
			shouldMatch:    true,
			expectedParams: map[string]string{},
		},
		{
			pattern:     "/id/{id:[0-9]+}",
			path:        "/id/123",
			shouldMatch: true,
			expectedParams: map[string]string{
				"id": "123",
			},
		},
		{
			pattern:     "/id/{id:[0-9]+}",
			path:        "/id/abc",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+" -> "+tt.path, func(t *testing.T) {
			matcher, err := CompilePattern(tt.pattern)
			require.NoError(t, err)

			params, matched := matcher.Match(tt.path)
			assert.Equal(t, tt.shouldMatch, matched)
			if tt.shouldMatch && tt.expectedParams != nil {
				assert.Equal(t, tt.expectedParams, params)
			}
		})
	}
}

func TestCompilePatternError(t *testing.T) {
	_, err := CompilePattern("/users/{id}/{id}")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate parameter name")
}
