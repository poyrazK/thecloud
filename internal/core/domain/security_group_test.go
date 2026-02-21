package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const (
	testSGName = "test-sg"
	anyIPv4    = "0.0.0.0/0"
)

func TestSecurityGroupValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		sg      SecurityGroup
		wantErr bool
		msg     string
	}{
		{
			name: "valid security group",
			sg: SecurityGroup{
				Name:   testSGName,
				VPCID:  uuid.New(),
				UserID: uuid.New(),
			},
			wantErr: false,
		},
		{
			name: "empty name",
			sg: SecurityGroup{
				Name:   "",
				VPCID:  uuid.New(),
				UserID: uuid.New(),
			},
			wantErr: true,
			msg:     "security group name cannot be empty",
		},
		{
			name: "nil vpc id",
			sg: SecurityGroup{
				Name:   testSGName,
				VPCID:  uuid.Nil,
				UserID: uuid.New(),
			},
			wantErr: true,
			msg:     "security group must be associated with a VPC",
		},
		{
			name: "nil user id",
			sg: SecurityGroup{
				Name:   "test-sg",
				VPCID:  uuid.New(),
				UserID: uuid.Nil,
			},
			wantErr: true,
			msg:     "security group must have a user owner",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.sg.Validate()
			if tt.wantErr {
				require.Error(t, err)
				if tt.msg != "" {
					assert.Contains(t, err.Error(), tt.msg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSecurityRuleValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		rule    SecurityRule
		wantErr bool
		msg     string
	}{
		{
			name: "valid tcp rule",
			rule: SecurityRule{
				Direction: RuleIngress,
				Protocol:  "tcp",
				PortMin:   80,
				PortMax:   80,
				CIDR:      anyIPv4,
			},
			wantErr: false,
		},
		{
			name: "valid udp range",
			rule: SecurityRule{
				Direction: RuleEgress,
				Protocol:  "udp",
				PortMin:   1000,
				PortMax:   2000,
				CIDR:      "10.0.0.0/8",
			},
			wantErr: false,
		},
		{
			name: "valid icmp",
			rule: SecurityRule{
				Direction: RuleIngress,
				Protocol:  "icmp",
				CIDR:      anyIPv4,
			},
			wantErr: false,
		},
		{
			name: "valid all",
			rule: SecurityRule{
				Direction: RuleIngress,
				Protocol:  "all",
				CIDR:      anyIPv4,
			},
			wantErr: false,
		},
		{
			name: "invalid direction",
			rule: SecurityRule{
				Direction: "invalid",
				Protocol:  "tcp",
				CIDR:      anyIPv4,
			},
			wantErr: true,
			msg:     "invalid rule direction",
		},
		{
			name: "invalid protocol",
			rule: SecurityRule{
				Direction: RuleIngress,
				Protocol:  "invalid",
				CIDR:      anyIPv4,
			},
			wantErr: true,
			msg:     "invalid protocol",
		},
		{
			name: "invalid port range",
			rule: SecurityRule{
				Direction: RuleIngress,
				Protocol:  "tcp",
				PortMin:   443,
				PortMax:   80,
				CIDR:      anyIPv4,
			},
			wantErr: true,
			msg:     "port_min (443) cannot be greater than port_max (80)",
		},
		{
			name: "port_min out of range",
			rule: SecurityRule{
				Direction: RuleIngress,
				Protocol:  "tcp",
				PortMin:   0,
				PortMax:   80,
				CIDR:      anyIPv4,
			},
			wantErr: true,
			msg:     "invalid port_min",
		},
		{
			name: "port_max out of range",
			rule: SecurityRule{
				Direction: RuleIngress,
				Protocol:  "tcp",
				PortMin:   80,
				PortMax:   70000,
				CIDR:      anyIPv4,
			},
			wantErr: true,
			msg:     "invalid port_max",
		},
		{
			name: "invalid cidr",
			rule: SecurityRule{
				Direction: RuleIngress,
				Protocol:  "tcp",
				PortMin:   80,
				PortMax:   80,
				CIDR:      "invalid-cidr",
			},
			wantErr: true,
			msg:     "invalid CIDR",
		},
		{
			name: "empty cidr",
			rule: SecurityRule{
				Direction: RuleIngress,
				Protocol:  "tcp",
				PortMin:   80,
				PortMax:   80,
				CIDR:      "",
			},
			wantErr: true,
			msg:     "CIDR is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.Validate()
			if tt.wantErr {
				require.Error(t, err)
				if tt.msg != "" {
					assert.Contains(t, err.Error(), tt.msg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
