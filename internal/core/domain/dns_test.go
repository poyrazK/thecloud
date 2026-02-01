package domain_test

import (
	"testing"

	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

func TestValidRecordTypes(t *testing.T) {
	types := domain.ValidRecordTypes()
	assert.NotEmpty(t, types)
	assert.Contains(t, types, domain.RecordTypeA)
	assert.Contains(t, types, domain.RecordTypeAAAA)
	assert.Contains(t, types, domain.RecordTypeCNAME)
	assert.Contains(t, types, domain.RecordTypeMX)
	assert.Contains(t, types, domain.RecordTypeTXT)
	assert.Contains(t, types, domain.RecordTypeSRV)
}

func TestIsValidRecordType(t *testing.T) {
	tests := []struct {
		name     string
		record   domain.RecordType
		expected bool
	}{
		{"A", domain.RecordTypeA, true},
		{"AAAA", domain.RecordTypeAAAA, true},
		{"CNAME", domain.RecordTypeCNAME, true},
		{"MX", domain.RecordTypeMX, true},
		{"TXT", domain.RecordTypeTXT, true},
		{"SRV", domain.RecordTypeSRV, true},
		{"INVALID", "INVALID", false},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, domain.IsValidRecordType(tt.record))
		})
	}
}
