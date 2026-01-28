// Package httphandlers exposes HTTP handlers for the API.
package httphandlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/poyrazk/thecloud/pkg/httputil"
)

const (
	errInvalidClusterID = "invalid cluster id"
	errInvalidRequest   = "invalid request body"
)

// ClusterHandler handles managed Kubernetes HTTP endpoints.
type ClusterHandler struct {
	svc ports.ClusterService
}

// NewClusterHandler constructs a new ClusterHandler.
func NewClusterHandler(svc ports.ClusterService) *ClusterHandler {
	return &ClusterHandler{svc: svc}
}

// CreateClusterRequest is the payload for creating a K8s cluster.
type CreateClusterRequest struct {
	Name             string `json:"name" binding:"required"`
	VpcID            string `json:"vpc_id" binding:"required"`
	Version          string `json:"version"`
	Workers          int    `json:"workers"`
	NetworkIsolation bool   `json:"network_isolation"`
	HA               bool   `json:"ha"`
}

// CreateCluster godoc
// @Summary Create a managed K8s cluster
// @Description Provisions a new Kubernetes cluster using kubeadm
// @Tags K8s
// @Accept json
// @Produce json
// @Param request body CreateClusterRequest true "Cluster details"
// @Success 202 {object} domain.Cluster
// @Router /clusters [post]
func (h *ClusterHandler) CreateCluster(c *gin.Context) {
	var req CreateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, errInvalidRequest))
		return
	}

	vpcID, err := uuid.Parse(req.VpcID)
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid vpc_id"))
		return
	}

	userID := appcontext.UserIDFromContext(c.Request.Context())

	// Default values are handled in the service layer
	cluster, err := h.svc.CreateCluster(c.Request.Context(), ports.CreateClusterParams{
		UserID:           userID,
		Name:             req.Name,
		VpcID:            vpcID,
		Version:          req.Version,
		Workers:          req.Workers,
		NetworkIsolation: req.NetworkIsolation,
		HAEnabled:        req.HA,
	})
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusAccepted, cluster)
}

// GetCluster godoc
// @Summary Get cluster details
// @Description Returns cluster metadata and current status
// @Tags K8s
// @Produce json
// @Param id path string true "Cluster ID"
// @Success 200 {object} domain.Cluster
// @Router /clusters/{id} [get]
func (h *ClusterHandler) GetCluster(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, errInvalidClusterID))
		return
	}

	cluster, err := h.svc.GetCluster(c.Request.Context(), id)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, cluster)
}

// ListClusters godoc
// @Summary List managed K8s clusters
// @Description Returns all clusters belonging to the user
// @Tags K8s
// @Produce json
// @Success 200 {array} domain.Cluster
// @Router /clusters [get]
func (h *ClusterHandler) ListClusters(c *gin.Context) {
	userID := appcontext.UserIDFromContext(c.Request.Context())
	clusters, err := h.svc.ListClusters(c.Request.Context(), userID)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, clusters)
}

// DeleteCluster godoc
// @Summary Delete a K8s cluster
// @Description Terminates all nodes and removes the cluster record
// @Tags K8s
// @Param id path string true "Cluster ID"
// @Success 202
// @Router /clusters/{id} [delete]
func (h *ClusterHandler) DeleteCluster(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, errInvalidClusterID))
		return
	}

	if err := h.svc.DeleteCluster(c.Request.Context(), id); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusAccepted, nil)
}

// GetKubeconfig godoc
// @Summary Download kubeconfig
// @Description Returns the kubeconfig for clinical access to the cluster
// @Tags K8s
// @Produce plain
// @Param id path string true "Cluster ID"
// @Param role query string false "Role (e.g. viewer)"
// @Success 200 {string} string
// @Router /clusters/{id}/kubeconfig [get]
func (h *ClusterHandler) GetKubeconfig(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, errInvalidClusterID))
		return
	}

	role := c.Query("role")
	kubeconfig, err := h.svc.GetKubeconfig(c.Request.Context(), id, role)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, kubeconfig)
}

// RepairCluster godoc
// @Summary Repair cluster components
// @Description Re-applies CNI and kube-proxy patches to a running cluster
// @Tags K8s
// @Param id path string true "Cluster ID"
// @Success 202
// @Router /clusters/{id}/repair [post]
func (h *ClusterHandler) RepairCluster(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, errInvalidClusterID))
		return
	}

	if err := h.svc.RepairCluster(c.Request.Context(), id); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusAccepted, nil)
}

// ScaleClusterRequest is the payload for scaling workers.
type ScaleClusterRequest struct {
	Workers int `json:"workers"`
}

// ScaleCluster godoc
// @Summary Scale cluster workers
// @Description Adjusts the number of worker nodes
// @Tags K8s
// @Param id path string true "Cluster ID"
// @Param request body ScaleClusterRequest true "Scale Request"
// @Success 200
// @Router /clusters/{id}/scale [post]
func (h *ClusterHandler) ScaleCluster(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, errInvalidClusterID))
		return
	}

	var req ScaleClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, errInvalidRequest))
		return
	}

	if err := h.svc.ScaleCluster(c.Request.Context(), id, req.Workers); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, nil)
}

// GetClusterHealth godoc
// @Summary Get cluster operational health
// @Description Returns readiness of nodes and API server
// @Tags K8s
// @Produce json
// @Param id path string true "Cluster ID"
// @Success 200 {object} ports.ClusterHealth
// @Router /clusters/{id}/health [get]
func (h *ClusterHandler) GetClusterHealth(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, errInvalidClusterID))
		return
	}

	health, err := h.svc.GetClusterHealth(c.Request.Context(), id)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, health)
}

// UpgradeClusterRequest is the payload for cluster upgrade.
type UpgradeClusterRequest struct {
	Version string `json:"version" binding:"required"`
}

// UpgradeCluster godoc
// @Summary Upgrade cluster version
// @Description Initiates an asynchronous upgrade of the Kubernetes control plane and workers
// @Tags K8s
// @Param id path string true "Cluster ID"
// @Param request body UpgradeClusterRequest true "Upgrade Request"
// @Success 202
// @Router /clusters/{id}/upgrade [post]
func (h *ClusterHandler) UpgradeCluster(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, errInvalidClusterID))
		return
	}

	var req UpgradeClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid version"))
		return
	}

	if err := h.svc.UpgradeCluster(c.Request.Context(), id, req.Version); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusAccepted, nil)
}

// RotateSecrets godoc
// @Summary Rotate cluster secrets
// @Description Renews cluster certificates and refreshes admin kubeconfig
// @Tags K8s
// @Param id path string true "Cluster ID"
// @Success 200
// @Router /clusters/{id}/rotate-secrets [post]
func (h *ClusterHandler) RotateSecrets(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, errInvalidClusterID))
		return
	}

	if err := h.svc.RotateSecrets(c.Request.Context(), id); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, nil)
}

// CreateBackup godoc
// @Summary Create cluster backup
// @Description Creates an etcd snapshot of the cluster state
// @Tags K8s
// @Param id path string true "Cluster ID"
// @Success 202
// @Router /clusters/{id}/backups [post]
func (h *ClusterHandler) CreateBackup(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, errInvalidClusterID))
		return
	}

	if err := h.svc.CreateBackup(c.Request.Context(), id); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusAccepted, nil)
}

// RestoreBackupRequest is the payload for cluster restoration.
type RestoreBackupRequest struct {
	BackupPath string `json:"backup_path" binding:"required"`
}

// RestoreBackup godoc
// @Summary Restore cluster from backup
// @Description Restores the etcd state from a specified snapshot path
// @Tags K8s
// @Param id path string true "Cluster ID"
// @Param request body RestoreBackupRequest true "Restore Request"
// @Success 200
// @Router /clusters/{id}/restore [post]
func (h *ClusterHandler) RestoreBackup(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, errInvalidClusterID))
		return
	}

	var req RestoreBackupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid backup path"))
		return
	}

	if err := h.svc.RestoreBackup(c.Request.Context(), id, req.BackupPath); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, nil)
}
