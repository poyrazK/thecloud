package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestElasticIP_Validate(t *testing.T) {
	tests := []struct {
		name    string
		eip     *ElasticIP
		wantErr bool
		msg     string
	}{
		{
			name: "valid",
			eip: &ElasticIP{
				UserID:   uuid.New(),
				TenantID: uuid.New(),
				PublicIP: "1.2.3.4",
				Status:   EIPStatusAllocated,
			},
			wantErr: false,
		},
		{
			name: "missing user id",
			eip: &ElasticIP{
				TenantID: uuid.New(),
				PublicIP: "1.2.3.4",
				Status:   EIPStatusAllocated,
			},
			wantErr: true,
			msg:     "user ID is required",
		},
		{
			name: "missing tenant id",
			eip: &ElasticIP{
				UserID:   uuid.New(),
				PublicIP: "1.2.3.4",
				Status:   EIPStatusAllocated,
			},
			wantErr: true,
			msg:     "tenant ID is required",
		},
		{
			name: "missing ip",
			eip: &ElasticIP{
				UserID:   uuid.New(),
				TenantID: uuid.New(),
				Status:   EIPStatusAllocated,
			},
			wantErr: true,
			msg:     "public IP address is required",
		},
		{
			name: "missing status",
			eip: &ElasticIP{
				UserID:   uuid.New(),
				TenantID: uuid.New(),
				PublicIP: "1.2.3.4",
			},
			wantErr: true,
			msg:     "status is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.eip.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, tt.msg, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
