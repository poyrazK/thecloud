package ratelimit

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestGetLimiterReturnsSingletonPerKey(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Limit(1), 1, slog.New(slog.NewTextHandler(io.Discard, nil)))
	first := limiter.GetLimiter("ip-1")
	second := limiter.GetLimiter("ip-1")

	assert.Equal(t, first, second)
}

func newTestContext(method string, path string, ip string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, nil)
	c.Request.RemoteAddr = ip + ":1234"
	return c, w
}

func TestMiddlewareAllowsFirstRequest(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Limit(1), 1, slog.New(slog.NewTextHandler(io.Discard, nil)))
	ctx, resp := newTestContext("GET", "/", testutil.TestNoopIP3)

	Middleware(limiter)(ctx)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestMiddlewareBlocksWhenRateExceeded(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Limit(1), 1, slog.New(slog.NewTextHandler(io.Discard, nil)))
	ctx, _ := newTestContext("GET", "/", testutil.TestNoopIP4)
	Middleware(limiter)(ctx)

	ctx2, resp2 := newTestContext("GET", "/", testutil.TestNoopIP4)
	Middleware(limiter)(ctx2)

	assert.Equal(t, http.StatusTooManyRequests, resp2.Code)
}
