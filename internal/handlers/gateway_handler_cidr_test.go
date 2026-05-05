package httphandlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

func TestCheckCIDR(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		remoteAddr     string
		blockedCIDRs   []string
		allowedCIDRs   []string
		expectedResult bool
	}{
		{
			name:           "no restrictions should allow all",
			remoteAddr:     "10.0.0.1:12345",
			blockedCIDRs:   []string{},
			allowedCIDRs:   []string{},
			expectedResult: true,
		},
		{
			name:           "empty CIDR lists should allow all",
			remoteAddr:     "10.0.0.1:12345",
			blockedCIDRs:   nil,
			allowedCIDRs:   nil,
			expectedResult: true,
		},
		{
			name:           "invalid IP should be denied (fail closed)",
			remoteAddr:     "invalid-ip:12345",
			blockedCIDRs:   nil,
			allowedCIDRs:   nil,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)
			handler := &GatewayHandler{logger: nil}
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest("GET", "/", nil)
			if tt.remoteAddr != "" {
				c.Request.RemoteAddr = tt.remoteAddr
			}

			route := &domain.GatewayRoute{
				BlockedCIDRs: tt.blockedCIDRs,
				AllowedCIDRs: tt.allowedCIDRs,
			}

			result := handler.checkCIDR(c, route)
			assert.Equal(t, tt.expectedResult, result, tt.name)
		})
	}
}