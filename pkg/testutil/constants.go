// Package testutil provides shared constants and helpers for testing.
package testutil

const (
	// Network related constants
	TestIPLocalhost       = "127.0.0.1"
	TestIPHost            = "10.0.0.1"
	TestIPPrivate         = "192.168.1.1"
	TestCIDR              = "10.0.0.0/16"
	TestOtherCIDR         = "192.168.1.0/24"
	TestAnyCIDR           = "0.0.0.0/0"
	TestGoogleDNSCIDR     = "8.8.8.8/32"
	TestTenCIDR           = "10.0.0.0/8"
	TestSubnetCIDR        = "10.0.1.0/24"
	TestGatewayIP         = "10.0.1.1"
	TestInstanceIP        = "10.0.1.2"
	TestDockerInstanceIP  = "10.0.0.2"
	TestNoopIP1           = "1.1.1.1"
	TestNoopIP2           = "1.1.1.2"
	TestNoopIP3           = "1.2.3.4"
	TestNoopIP4           = "2.2.2.2"
	TestLibvirtInstanceIP = "192.168.122.10"
	TestLibvirtDHCPStart  = "10.0.0.2"
	TestLibvirtDHCPEnd    = "10.0.0.50"
	TestLibvirtPoolStart  = "192.168.100.0"
	TestLibvirtPoolEnd    = "192.168.200.255"
	TestVXLANRemoteIP     = "192.168.1.1"
	TestSubnet2CIDR       = "10.0.2.0/24"

	// Auth related constants
	TestPasswordStrong     = "CorrectHorseBatteryStaple123!"
	TestPasswordWeak       = "123"
	TestEmail              = "test@example.com"
	TestUserAgent          = "test-agent"
	TestBaseURL            = "http://localhost:8080"
	TestHeaderAPIKey       = "X-API-Key"
	TestContentTypeAppJSON = "application/json"
	TestDatabaseURL        = "postgres://cloud:cloud@localhost:5433/thecloud"
	TestProdDatabaseURL    = "postgres://test:test@localhost:5432/testdb"
	TestEnvDev             = "development"
	TestEnvProd            = "production"
	TestPort               = "8080"
	TestProdPort           = "9090"
)
