package services

import (
	"fmt"
	"strings"
	"testing"

	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/stretchr/testify/assert"
)

func TestParseAndValidatePortsTooMany(t *testing.T) {
	t.Parallel()
	svc := &InstanceService{}
	var builder strings.Builder
	for i := 0; i < domain.MaxPortsPerInstance+1; i++ {
		if i > 0 {
			builder.WriteString(",")
		}
		builder.WriteString(fmt.Sprintf("%d:%d", 8000+i, 80+i))
	}

	_, err := svc.parseAndValidatePorts(builder.String())
	assert.Error(t, err)
	assert.True(t, errors.Is(err, errors.TooManyPorts))
}

func TestValidatePortMappingFormatErrors(t *testing.T) {
	t.Parallel()
	tests := []string{
		"",          // missing colon
		"8080",      // no mapping
		"8080:80:1", // too many colon
		"foo:80",    // non-numeric host
		"80:bar",    // non-numeric container
		"65536:80",  // host out of range
		"80:65536",  // container out of range
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			err := validatePortMapping(input)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, errors.InvalidPortFormat))
		})
	}
}

func TestParsePortEmptyFails(t *testing.T) {
	t.Parallel()
	_, err := parsePort("   ")
	assert.Error(t, err)
}
