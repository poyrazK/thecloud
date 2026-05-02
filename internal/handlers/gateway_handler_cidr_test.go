package httphandlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

func TestCheckCIDR_NoBlockedOrAllowed(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	handler := &GatewayHandler{logger: nil}
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.RemoteAddr = "10.0.0.1:12345"

	route := &domain.GatewayRoute{
		BlockedCIDRs: []string{},
		AllowedCIDRs: []string{},
	}

	result := handler.checkCIDR(c, route)
	assert.True(t, result, "no restrictions should allow all")
}

func TestCheckCIDR_EmptyBlockedAndAllowed(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	handler := &GatewayHandler{logger: nil}
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.RemoteAddr = "10.0.0.1:12345"

	route := &domain.GatewayRoute{}

	result := handler.checkCIDR(c, route)
	assert.True(t, result, "empty CIDR lists should allow all")
}

func TestCheckCIDR_InvalidIP(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	handler := &GatewayHandler{logger: nil}
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/", nil)
	// Leave RemoteAddr empty to test invalid IP handling

	route := &domain.GatewayRoute{}

	result := handler.checkCIDR(c, route)
	assert.True(t, result, "invalid IP should be allowed (fail open)")
}
