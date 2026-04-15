package services_test

import (
	"testing"

	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/mock"
	testifyMock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupGlobalLBTest(t *testing.T) (*services.GlobalLBService, *mock.MockGlobalLBRepo, *mock.MockLBRepo, *mock.MockGeoDNS) {
	t.Helper()
	repo := mock.NewMockGlobalLBRepo()
	lbRepo := mock.NewMockLBRepo()
	geoDNS := mock.NewMockGeoDNS()
	audit := mock.NewMockAuditService()
	logger := mock.NewNoopLogger()

	rbacSvc := new(MockRBACService)
	rbacSvc.On("Authorize", testifyMock.Anything, testifyMock.Anything, testifyMock.Anything, testifyMock.Anything, testifyMock.Anything).Return(nil)

	svc := services.NewGlobalLBService(services.GlobalLBServiceParams{
		Repo: repo, RBAC: rbacSvc, LBRepo: lbRepo, GeoDNS: geoDNS, AuditSvc: audit, Logger: logger,
	})
	mockGeoDNS, ok := geoDNS.(*mock.MockGeoDNS)
	require.True(t, ok)
	return svc, repo, lbRepo, mockGeoDNS
}

