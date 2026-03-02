// Package httphandlers provides HTTP handlers for the API.
package httphandlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/poyrazk/thecloud/pkg/httputil"
)

const invalidDatabaseIDMsg = "invalid database id"

// DatabaseHandler handles database HTTP endpoints.
type DatabaseHandler struct {
	svc ports.DatabaseService
}

// NewDatabaseHandler constructs a DatabaseHandler.
func NewDatabaseHandler(svc ports.DatabaseService) *DatabaseHandler {
	return &DatabaseHandler{svc: svc}
}

// CreateDatabaseRequest is the payload for database creation.
type CreateDatabaseRequest struct {
	Name             string            `json:"name" binding:"required"`
	Engine           string            `json:"engine" binding:"required"`
	Version          string            `json:"version" binding:"required"`
	VpcID            *uuid.UUID        `json:"vpc_id"`
	AllocatedStorage int               `json:"allocated_storage"`
	Parameters       map[string]string `json:"parameters"`
	MetricsEnabled   bool              `json:"metrics_enabled"`
	PoolingEnabled   bool              `json:"pooling_enabled"`
}

func (h *DatabaseHandler) Create(c *gin.Context) {
	var req CreateDatabaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, err.Error()))
		return
	}

	db, err := h.svc.CreateDatabase(c.Request.Context(), ports.CreateDatabaseRequest{
		Name:             req.Name,
		Engine:           req.Engine,
		Version:          req.Version,
		VpcID:            req.VpcID,
		AllocatedStorage: req.AllocatedStorage,
		Parameters:       req.Parameters,
		MetricsEnabled:   req.MetricsEnabled,
		PoolingEnabled:   req.PoolingEnabled,
	})
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusCreated, db)
}

// CreateReplicaRequest is the payload for creating a database replica.
type CreateReplicaRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *DatabaseHandler) CreateReplica(c *gin.Context) {
	primaryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, invalidDatabaseIDMsg))
		return
	}

	var req CreateReplicaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, err.Error()))
		return
	}

	db, err := h.svc.CreateReplica(c.Request.Context(), primaryID, req.Name)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusCreated, db)
}

func (h *DatabaseHandler) Promote(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, invalidDatabaseIDMsg))
		return
	}

	if err := h.svc.PromoteToPrimary(c.Request.Context(), id); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{"message": "database promoted to primary"})
}

func (h *DatabaseHandler) List(c *gin.Context) {
	dbs, err := h.svc.ListDatabases(c.Request.Context())
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, dbs)
}

func (h *DatabaseHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, invalidDatabaseIDMsg))
		return
	}

	db, err := h.svc.GetDatabase(c.Request.Context(), id)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, db)
}

func (h *DatabaseHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, invalidDatabaseIDMsg))
		return
	}

	if err := h.svc.DeleteDatabase(c.Request.Context(), id); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{"message": "database deleted"})
}

func (h *DatabaseHandler) GetConnectionString(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, invalidDatabaseIDMsg))
		return
	}

	connStr, err := h.svc.GetConnectionString(c.Request.Context(), id)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{"connection_string": connStr})
}

// CreateDatabaseSnapshotRequest is the payload for creating a database snapshot.
type CreateDatabaseSnapshotRequest struct {
	Description string `json:"description"`
}

// RestoreDatabaseRequest is the payload for restoring a database from a snapshot.
type RestoreDatabaseRequest struct {
	SnapshotID       uuid.UUID         `json:"snapshot_id" binding:"required"`
	Name             string            `json:"name" binding:"required"`
	Engine           string            `json:"engine" binding:"required"`
	Version          string            `json:"version" binding:"required"`
	VpcID            *uuid.UUID        `json:"vpc_id"`
	AllocatedStorage int               `json:"allocated_storage" binding:"required"`
	Parameters       map[string]string `json:"parameters"`
	MetricsEnabled   bool              `json:"metrics_enabled"`
	PoolingEnabled   bool              `json:"pooling_enabled"`
}

// CreateSnapshot creates a point-in-time backup of the database.
// @Summary Create database snapshot
// @Description Creates a point-in-time backup of the database underlying volume
// @Tags databases
// @Accept json
// @Produce json
// @Security APIKeyAuth
// @Param id path string true "Database ID"
// @Param request body CreateDatabaseSnapshotRequest false "Snapshot description"
// @Success 201 {object} domain.Snapshot
// @Router /databases/{id}/snapshots [post]
func (h *DatabaseHandler) CreateSnapshot(c *gin.Context) {
	databaseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, invalidDatabaseIDMsg))
		return
	}

	var req CreateDatabaseSnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" { // Allow empty body
		httputil.Error(c, errors.New(errors.InvalidInput, err.Error()))
		return
	}

	snap, err := h.svc.CreateDatabaseSnapshot(c.Request.Context(), databaseID, req.Description)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusCreated, snap)
}

// ListSnapshots returns all snapshots for a database.
// @Summary List database snapshots
// @Description Returns all snapshots belonging to a specific database
// @Tags databases
// @Produce json
// @Security APIKeyAuth
// @Param id path string true "Database ID"
// @Success 200 {array} domain.Snapshot
// @Router /databases/{id}/snapshots [get]
func (h *DatabaseHandler) ListSnapshots(c *gin.Context) {
	databaseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, invalidDatabaseIDMsg))
		return
	}

	snaps, err := h.svc.ListDatabaseSnapshots(c.Request.Context(), databaseID)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, snaps)
}

// Restore provisions a new database from a snapshot.
// @Summary Restore database from snapshot
// @Description Creates a new database instance from an existing volume snapshot
// @Tags databases
// @Accept json
// @Produce json
// @Security APIKeyAuth
// @Param request body RestoreDatabaseRequest true "Restore parameters"
// @Success 201 {object} domain.Database
// @Router /databases/restore [post]
func (h *DatabaseHandler) Restore(c *gin.Context) {
	var req RestoreDatabaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, err.Error()))
		return
	}

	db, err := h.svc.RestoreDatabase(c.Request.Context(), ports.RestoreDatabaseRequest{
		SnapshotID:       req.SnapshotID,
		NewName:          req.Name,
		Engine:           req.Engine,
		Version:          req.Version,
		VpcID:            req.VpcID,
		AllocatedStorage: req.AllocatedStorage,
		Parameters:       req.Parameters,
		MetricsEnabled:   req.MetricsEnabled,
		PoolingEnabled:   req.PoolingEnabled,
	})
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusCreated, db)
}
