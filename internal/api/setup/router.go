// Package setup wires API dependencies and routes.
package setup

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	_ "github.com/poyrazk/thecloud/docs/swagger"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	httphandlers "github.com/poyrazk/thecloud/internal/handlers"
	"github.com/poyrazk/thecloud/internal/handlers/ws"
	"github.com/poyrazk/thecloud/internal/platform"
	"github.com/poyrazk/thecloud/pkg/httputil"
	"github.com/poyrazk/thecloud/pkg/ratelimit"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/time/rate"
)

const (
	bucketKeyRoute = "/:bucket/*key"
	roleIDRoute    = "/roles/:id"
)

// Handlers bundles HTTP handlers used by the router.
type Handlers struct {
	Audit         *httphandlers.AuditHandler
	Identity      *httphandlers.IdentityHandler
	Auth          *httphandlers.AuthHandler
	Vpc           *httphandlers.VpcHandler
	Subnet        *httphandlers.SubnetHandler
	Instance      *httphandlers.InstanceHandler
	Event         *httphandlers.EventHandler
	Volume        *httphandlers.VolumeHandler
	LB            *httphandlers.LBHandler
	Dashboard     *httphandlers.DashboardHandler
	RBAC          *httphandlers.RBACHandler
	Snapshot      *httphandlers.SnapshotHandler
	Stack         *httphandlers.StackHandler
	Storage       *httphandlers.StorageHandler
	Database      *httphandlers.DatabaseHandler
	Secret        *httphandlers.SecretHandler
	Function      *httphandlers.FunctionHandler
	Cache         *httphandlers.CacheHandler
	Queue         *httphandlers.QueueHandler
	Notify        *httphandlers.NotifyHandler
	Cron          *httphandlers.CronHandler
	Gateway       *httphandlers.GatewayHandler
	Container     *httphandlers.ContainerHandler
	Health        *httphandlers.HealthHandler
	SecurityGroup *httphandlers.SecurityGroupHandler
	AutoScaling   *httphandlers.AutoScalingHandler
	Accounting    *httphandlers.AccountingHandler
	Image         *httphandlers.ImageHandler
	Cluster       *httphandlers.ClusterHandler
	Lifecycle     *httphandlers.LifecycleHandler
	Ws            *ws.Handler
}

// InitHandlers constructs HTTP handlers and websocket hub.
func InitHandlers(svcs *Services, cfg *platform.Config, logger *slog.Logger) *Handlers {
	hub := ws.NewHub(logger)
	go hub.Run()

	return &Handlers{
		Audit:         httphandlers.NewAuditHandler(svcs.Audit),
		Identity:      httphandlers.NewIdentityHandler(svcs.Identity),
		Auth:          httphandlers.NewAuthHandler(svcs.Auth, svcs.PasswordReset),
		Vpc:           httphandlers.NewVpcHandler(svcs.Vpc),
		Subnet:        httphandlers.NewSubnetHandler(svcs.Subnet),
		Instance:      httphandlers.NewInstanceHandler(svcs.Instance),
		Event:         httphandlers.NewEventHandler(svcs.Event),
		Volume:        httphandlers.NewVolumeHandler(svcs.Volume),
		LB:            httphandlers.NewLBHandler(svcs.LB),
		Dashboard:     httphandlers.NewDashboardHandler(svcs.Dashboard),
		RBAC:          httphandlers.NewRBACHandler(svcs.RBAC),
		Snapshot:      httphandlers.NewSnapshotHandler(svcs.Snapshot),
		Stack:         httphandlers.NewStackHandler(svcs.Stack),
		Storage:       httphandlers.NewStorageHandler(svcs.Storage, cfg),
		Database:      httphandlers.NewDatabaseHandler(svcs.Database),
		Secret:        httphandlers.NewSecretHandler(svcs.Secret),
		Function:      httphandlers.NewFunctionHandler(svcs.Function),
		Cache:         httphandlers.NewCacheHandler(svcs.Cache),
		Queue:         httphandlers.NewQueueHandler(svcs.Queue),
		Notify:        httphandlers.NewNotifyHandler(svcs.Notify),
		Cron:          httphandlers.NewCronHandler(svcs.Cron),
		Gateway:       httphandlers.NewGatewayHandler(svcs.Gateway),
		Container:     httphandlers.NewContainerHandler(svcs.Container),
		Health:        httphandlers.NewHealthHandler(svcs.Health),
		SecurityGroup: httphandlers.NewSecurityGroupHandler(svcs.SecurityGroup),
		AutoScaling:   httphandlers.NewAutoScalingHandler(svcs.AutoScaling),
		Accounting:    httphandlers.NewAccountingHandler(svcs.Accounting),
		Image:         httphandlers.NewImageHandler(svcs.Image),
		Cluster:       httphandlers.NewClusterHandler(svcs.Cluster),
		Lifecycle:     httphandlers.NewLifecycleHandler(svcs.Lifecycle),
		Ws:            ws.NewHandler(hub, svcs.Identity, logger),
	}
}

// SetupRouter wires all routes, middleware, and documentation endpoints.
func SetupRouter(cfg *platform.Config, logger *slog.Logger, handlers *Handlers, services *Services, networkBackend ports.NetworkBackend) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(httputil.RequestID())
	r.Use(httputil.Logger(logger))
	r.Use(httputil.CORS())
	r.Use(gin.Recovery())

	// Security Middleware
	r.Use(httputil.SecurityHeadersMiddleware())

	// Rate Limiter
	globalRPS, _ := strconv.Atoi(cfg.RateLimitGlobal)
	if globalRPS <= 0 {
		globalRPS = 5
	}
	limiter := ratelimit.NewIPRateLimiter(rate.Limit(globalRPS), globalRPS*2, logger)
	r.Use(ratelimit.Middleware(limiter))
	r.Use(httputil.Metrics())

	// 6. Routes
	r.GET("/health/live", handlers.Health.Live)
	r.GET("/health/ready", handlers.Health.Ready)
	r.GET("/health", handlers.Health.Ready)

	// OVS Health check
	r.GET("/health/ovs", func(c *gin.Context) {
		if networkBackend == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": "network backend not initialized"})
			return
		}

		// Check if using noop adapter (degraded mode)
		if networkBackend.Type() == "noop" {
			c.JSON(http.StatusOK, gin.H{"status": "degraded", "message": "using no-op network backend"})
			return
		}

		if err := networkBackend.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// 7. Profiling (pprof) - only in non-production
	if cfg.Environment != "production" {
		pprof.Register(r)
	}

	// Register Route Groups
	registerAuthRoutes(r, handlers, services, cfg, logger)
	registerComputeRoutes(r, handlers, services)
	registerNetworkRoutes(r, handlers, services)
	registerDataRoutes(r, handlers, services)
	registerDevOpsRoutes(r, handlers, services)
	registerAdminRoutes(r, handlers, services)

	// The actual Gateway Proxy (Public)
	r.Any("/gw/*proxy", handlers.Gateway.Proxy)

	return r
}

func registerAuthRoutes(r *gin.Engine, handlers *Handlers, svcs *Services, cfg *platform.Config, logger *slog.Logger) {
	authRPM, _ := strconv.Atoi(cfg.RateLimitAuth)
	if authRPM <= 0 {
		authRPM = 5
	}
	authLimiter := ratelimit.NewIPRateLimiter(rate.Limit(float64(authRPM)/60.0), authRPM, logger)
	authMiddleware := ratelimit.Middleware(authLimiter)

	r.POST("/auth/register", authMiddleware, handlers.Auth.Register)
	r.POST("/auth/login", authMiddleware, handlers.Auth.Login)
	r.POST("/auth/forgot-password", authMiddleware, handlers.Auth.ForgotPassword)
	r.POST("/auth/reset-password", authMiddleware, handlers.Auth.ResetPassword)

	keyGroup := r.Group("/auth/keys")
	keyGroup.Use(httputil.Auth(svcs.Identity))
	{
		keyGroup.POST("", handlers.Identity.CreateKey)
		keyGroup.GET("", handlers.Identity.ListKeys)
		keyGroup.DELETE("/:id", handlers.Identity.RevokeKey)
		keyGroup.POST("/:id/rotate", handlers.Identity.RotateKey)
		keyGroup.POST("/:id/regenerate", handlers.Identity.RegenerateKey)
	}
}

func registerComputeRoutes(r *gin.Engine, handlers *Handlers, svcs *Services) {
	instanceGroup := r.Group("/instances")
	instanceGroup.Use(httputil.Auth(svcs.Identity))
	{
		instanceGroup.POST("", httputil.Permission(svcs.RBAC, domain.PermissionInstanceLaunch), handlers.Instance.Launch)
		instanceGroup.GET("", httputil.Permission(svcs.RBAC, domain.PermissionInstanceRead), handlers.Instance.List)
		instanceGroup.GET("/:id", httputil.Permission(svcs.RBAC, domain.PermissionInstanceRead), handlers.Instance.Get)
		instanceGroup.POST("/:id/stop", httputil.Permission(svcs.RBAC, domain.PermissionInstanceUpdate), handlers.Instance.Stop)
		instanceGroup.GET("/:id/logs", httputil.Permission(svcs.RBAC, domain.PermissionInstanceRead), handlers.Instance.GetLogs)
		instanceGroup.GET("/:id/stats", httputil.Permission(svcs.RBAC, domain.PermissionInstanceRead), handlers.Instance.GetStats)
		instanceGroup.GET("/:id/console", httputil.Permission(svcs.RBAC, domain.PermissionInstanceRead), handlers.Instance.GetConsole)
		instanceGroup.DELETE("/:id", httputil.Permission(svcs.RBAC, domain.PermissionInstanceTerminate), handlers.Instance.Terminate)
	}

	snapshotGroup := r.Group("/snapshots")
	snapshotGroup.Use(httputil.Auth(svcs.Identity))
	{
		snapshotGroup.POST("", httputil.Permission(svcs.RBAC, domain.PermissionSnapshotCreate), handlers.Snapshot.Create)
		snapshotGroup.GET("", httputil.Permission(svcs.RBAC, domain.PermissionSnapshotRead), handlers.Snapshot.List)
		snapshotGroup.GET("/:id", httputil.Permission(svcs.RBAC, domain.PermissionSnapshotRead), handlers.Snapshot.Get)
		snapshotGroup.DELETE("/:id", httputil.Permission(svcs.RBAC, domain.PermissionSnapshotDelete), handlers.Snapshot.Delete)
		snapshotGroup.POST("/:id/restore", httputil.Permission(svcs.RBAC, domain.PermissionSnapshotRestore), handlers.Snapshot.Restore)
	}

	imageGroup := r.Group("/images")
	imageGroup.Use(httputil.Auth(svcs.Identity))
	{
		imageGroup.POST("", httputil.Permission(svcs.RBAC, domain.PermissionImageCreate), handlers.Image.RegisterImage)
		imageGroup.GET("", httputil.Permission(svcs.RBAC, domain.PermissionImageRead), handlers.Image.ListImages)
		imageGroup.GET("/:id", httputil.Permission(svcs.RBAC, domain.PermissionImageRead), handlers.Image.GetImage)
		imageGroup.DELETE("/:id", httputil.Permission(svcs.RBAC, domain.PermissionImageDelete), handlers.Image.DeleteImage)
		imageGroup.POST("/:id/upload", httputil.Permission(svcs.RBAC, domain.PermissionImageCreate), handlers.Image.UploadImage)
	}

	clusterGroup := r.Group("/clusters")
	clusterGroup.Use(httputil.Auth(svcs.Identity))
	{
		clusterGroup.POST("", httputil.Permission(svcs.RBAC, domain.PermissionClusterCreate), handlers.Cluster.CreateCluster)
		clusterGroup.GET("", httputil.Permission(svcs.RBAC, domain.PermissionClusterRead), handlers.Cluster.ListClusters)
		clusterGroup.GET("/:id", httputil.Permission(svcs.RBAC, domain.PermissionClusterRead), handlers.Cluster.GetCluster)
		clusterGroup.DELETE("/:id", httputil.Permission(svcs.RBAC, domain.PermissionClusterDelete), handlers.Cluster.DeleteCluster)
		clusterGroup.GET("/:id/kubeconfig", httputil.Permission(svcs.RBAC, domain.PermissionClusterRead), handlers.Cluster.GetKubeconfig)
		clusterGroup.POST("/:id/repair", httputil.Permission(svcs.RBAC, domain.PermissionClusterUpdate), handlers.Cluster.RepairCluster)
		clusterGroup.POST("/:id/scale", httputil.Permission(svcs.RBAC, domain.PermissionClusterUpdate), handlers.Cluster.ScaleCluster)
		clusterGroup.GET("/:id/health", httputil.Permission(svcs.RBAC, domain.PermissionClusterRead), handlers.Cluster.GetClusterHealth)
		clusterGroup.POST("/:id/upgrade", httputil.Permission(svcs.RBAC, domain.PermissionClusterUpdate), handlers.Cluster.UpgradeCluster)
		clusterGroup.POST("/:id/rotate-secrets", httputil.Permission(svcs.RBAC, domain.PermissionClusterUpdate), handlers.Cluster.RotateSecrets)
		clusterGroup.POST("/:id/backups", httputil.Permission(svcs.RBAC, domain.PermissionClusterUpdate), handlers.Cluster.CreateBackup)
		clusterGroup.POST("/:id/restore", httputil.Permission(svcs.RBAC, domain.PermissionClusterUpdate), handlers.Cluster.RestoreBackup)
	}
}

func registerNetworkRoutes(r *gin.Engine, handlers *Handlers, svcs *Services) {
	vpcGroup := r.Group("/vpcs")
	vpcGroup.Use(httputil.Auth(svcs.Identity))
	{
		vpcGroup.POST("", httputil.Permission(svcs.RBAC, domain.PermissionVpcCreate), handlers.Vpc.Create)
		vpcGroup.GET("", httputil.Permission(svcs.RBAC, domain.PermissionVpcRead), handlers.Vpc.List)
		vpcGroup.GET("/:id", httputil.Permission(svcs.RBAC, domain.PermissionVpcRead), handlers.Vpc.Get)
		vpcGroup.DELETE("/:id", httputil.Permission(svcs.RBAC, domain.PermissionVpcDelete), handlers.Vpc.Delete)

		vpcGroup.POST("/:id/subnets", httputil.Permission(svcs.RBAC, domain.PermissionVpcUpdate), handlers.Subnet.Create)
		vpcGroup.GET("/:id/subnets", httputil.Permission(svcs.RBAC, domain.PermissionVpcRead), handlers.Subnet.List)
	}

	subnetGroup := r.Group("/subnets")
	subnetGroup.Use(httputil.Auth(svcs.Identity))
	{
		subnetGroup.GET("/:id", httputil.Permission(svcs.RBAC, domain.PermissionVpcRead), handlers.Subnet.Get)
		subnetGroup.DELETE("/:id", httputil.Permission(svcs.RBAC, domain.PermissionVpcUpdate), handlers.Subnet.Delete)
	}

	sgGroup := r.Group("/security-groups")
	sgGroup.Use(httputil.Auth(svcs.Identity))
	{
		sgGroup.POST("", handlers.SecurityGroup.Create)
		sgGroup.GET("", handlers.SecurityGroup.List)
		sgGroup.GET("/:id", handlers.SecurityGroup.Get)
		sgGroup.DELETE("/:id", handlers.SecurityGroup.Delete)
		sgGroup.POST("/:id/rules", handlers.SecurityGroup.AddRule)
		sgGroup.DELETE("/rules/:rule_id", handlers.SecurityGroup.RemoveRule)
		sgGroup.POST("/attach", handlers.SecurityGroup.Attach)
		sgGroup.POST("/detach", handlers.SecurityGroup.Detach)
	}

	lbGroup := r.Group("/lb")
	lbGroup.Use(httputil.Auth(svcs.Identity))
	{
		lbGroup.POST("", httputil.Permission(svcs.RBAC, domain.PermissionLbCreate), handlers.LB.Create)
		lbGroup.GET("", httputil.Permission(svcs.RBAC, domain.PermissionLbRead), handlers.LB.List)
		lbGroup.GET("/:id", httputil.Permission(svcs.RBAC, domain.PermissionLbRead), handlers.LB.Get)
		lbGroup.DELETE("/:id", httputil.Permission(svcs.RBAC, domain.PermissionLbDelete), handlers.LB.Delete)
		lbGroup.POST("/:id/targets", httputil.Permission(svcs.RBAC, domain.PermissionLbUpdate), handlers.LB.AddTarget)
		lbGroup.GET("/:id/targets", httputil.Permission(svcs.RBAC, domain.PermissionLbRead), handlers.LB.ListTargets)
		lbGroup.DELETE("/:id/targets/:instanceId", httputil.Permission(svcs.RBAC, domain.PermissionLbUpdate), handlers.LB.RemoveTarget)
	}
}

func registerDataRoutes(r *gin.Engine, handlers *Handlers, svcs *Services) {
	storageGroup := r.Group("/storage")
	storageGroup.Use(httputil.Auth(svcs.Identity))
	{
		// explicitly registered static paths first
		storageGroup.GET("/cluster/status", handlers.Storage.GetClusterStatus)

		// Bucket Management (static/specific)
		storageGroup.POST("/buckets", handlers.Storage.CreateBucket)
		storageGroup.GET("/buckets", handlers.Storage.ListBuckets)
		storageGroup.DELETE("/buckets/:bucket", handlers.Storage.DeleteBucket)
		storageGroup.PATCH("/buckets/:bucket/versioning", handlers.Storage.SetBucketVersioning)

		// Lifecycle Management
		storageGroup.POST("/buckets/:bucket/lifecycle", handlers.Lifecycle.CreateRule)
		storageGroup.GET("/buckets/:bucket/lifecycle", handlers.Lifecycle.ListRules)
		storageGroup.DELETE("/buckets/:bucket/lifecycle/:id", handlers.Lifecycle.DeleteRule)

		// Versioning
		storageGroup.GET("/versions/:bucket/*key", handlers.Storage.ListVersions)

		// Multipart
		storageGroup.POST("/multipart/init/:bucket/*key", handlers.Storage.InitiateMultipartUpload)
		storageGroup.PUT("/multipart/upload/:id/parts", handlers.Storage.UploadPart)
		storageGroup.POST("/multipart/complete/:id", handlers.Storage.CompleteMultipartUpload)
		storageGroup.DELETE("/multipart/abort/:id", handlers.Storage.AbortMultipartUpload)

		// Presigned Generation (Auth Required)
		storageGroup.POST("/presign"+bucketKeyRoute, handlers.Storage.GeneratePresignedURL)

		// Parameterized object routes last
		storageGroup.PUT(bucketKeyRoute, handlers.Storage.Upload)
		storageGroup.GET(bucketKeyRoute, handlers.Storage.Download)
		storageGroup.DELETE(bucketKeyRoute, handlers.Storage.Delete)
		storageGroup.GET("/:bucket", handlers.Storage.List)
	}

	// Public Routes for Presigned Access (No Auth Middleware)
	r.GET("/storage/presigned"+bucketKeyRoute, handlers.Storage.ServePresignedDownload)
	r.PUT("/storage/presigned"+bucketKeyRoute, handlers.Storage.ServePresignedUpload)

	volumeGroup := r.Group("/volumes")
	volumeGroup.Use(httputil.Auth(svcs.Identity))
	{
		volumeGroup.POST("", httputil.Permission(svcs.RBAC, domain.PermissionVolumeCreate), handlers.Volume.Create)
		volumeGroup.GET("", httputil.Permission(svcs.RBAC, domain.PermissionVolumeRead), handlers.Volume.List)
		volumeGroup.GET("/:id", httputil.Permission(svcs.RBAC, domain.PermissionVolumeRead), handlers.Volume.Get)
		volumeGroup.DELETE("/:id", httputil.Permission(svcs.RBAC, domain.PermissionVolumeDelete), handlers.Volume.Delete)
	}

	dbGroup := r.Group("/databases")
	dbGroup.Use(httputil.Auth(svcs.Identity))
	{
		dbGroup.POST("", httputil.Permission(svcs.RBAC, domain.PermissionDBCreate), handlers.Database.Create)
		dbGroup.GET("", httputil.Permission(svcs.RBAC, domain.PermissionDBRead), handlers.Database.List)
		dbGroup.GET("/:id", httputil.Permission(svcs.RBAC, domain.PermissionDBRead), handlers.Database.Get)
		dbGroup.DELETE("/:id", httputil.Permission(svcs.RBAC, domain.PermissionDBDelete), handlers.Database.Delete)
		dbGroup.GET("/:id/connection", httputil.Permission(svcs.RBAC, domain.PermissionDBRead), handlers.Database.GetConnectionString)
	}

	cacheGroup := r.Group("/caches")
	cacheGroup.Use(httputil.Auth(svcs.Identity))
	{
		cacheGroup.POST("", httputil.Permission(svcs.RBAC, domain.PermissionCacheCreate), handlers.Cache.Create)
		cacheGroup.GET("", httputil.Permission(svcs.RBAC, domain.PermissionCacheRead), handlers.Cache.List)
		cacheGroup.GET("/:id", httputil.Permission(svcs.RBAC, domain.PermissionCacheRead), handlers.Cache.Get)
		cacheGroup.DELETE("/:id", httputil.Permission(svcs.RBAC, domain.PermissionCacheDelete), handlers.Cache.Delete)
		cacheGroup.GET("/:id/connection", httputil.Permission(svcs.RBAC, domain.PermissionCacheRead), handlers.Cache.GetConnectionString)
		cacheGroup.POST("/:id/flush", httputil.Permission(svcs.RBAC, domain.PermissionCacheUpdate), handlers.Cache.Flush)
		cacheGroup.GET("/:id/stats", httputil.Permission(svcs.RBAC, domain.PermissionCacheRead), handlers.Cache.GetStats)
	}

	secretGroup := r.Group("/secrets")
	secretGroup.Use(httputil.Auth(svcs.Identity))
	{
		secretGroup.POST("", httputil.Permission(svcs.RBAC, domain.PermissionSecretCreate), handlers.Secret.Create)
		secretGroup.GET("", httputil.Permission(svcs.RBAC, domain.PermissionSecretRead), handlers.Secret.List)
		secretGroup.GET("/:id", httputil.Permission(svcs.RBAC, domain.PermissionSecretRead), handlers.Secret.Get)
		secretGroup.DELETE("/:id", httputil.Permission(svcs.RBAC, domain.PermissionSecretDelete), handlers.Secret.Delete)
	}
}

func registerDevOpsRoutes(r *gin.Engine, handlers *Handlers, svcs *Services) {
	fnGroup := r.Group("/functions")
	fnGroup.Use(httputil.Auth(svcs.Identity))
	{
		fnGroup.POST("", httputil.Permission(svcs.RBAC, domain.PermissionFunctionCreate), handlers.Function.Create)
		fnGroup.GET("", httputil.Permission(svcs.RBAC, domain.PermissionFunctionRead), handlers.Function.List)
		fnGroup.GET("/:id", httputil.Permission(svcs.RBAC, domain.PermissionFunctionRead), handlers.Function.Get)
		fnGroup.DELETE("/:id", httputil.Permission(svcs.RBAC, domain.PermissionFunctionDelete), handlers.Function.Delete)
		fnGroup.POST("/:id/invoke", httputil.Permission(svcs.RBAC, domain.PermissionFunctionInvoke), handlers.Function.Invoke)
		fnGroup.GET("/:id/logs", httputil.Permission(svcs.RBAC, domain.PermissionFunctionRead), handlers.Function.GetLogs)
	}

	queueGroup := r.Group("/queues")
	queueGroup.Use(httputil.Auth(svcs.Identity))
	{
		queueGroup.POST("", httputil.Permission(svcs.RBAC, domain.PermissionQueueCreate), handlers.Queue.Create)
		queueGroup.GET("", httputil.Permission(svcs.RBAC, domain.PermissionQueueRead), handlers.Queue.List)
		queueGroup.GET("/:id", httputil.Permission(svcs.RBAC, domain.PermissionQueueRead), handlers.Queue.Get)
		queueGroup.DELETE("/:id", httputil.Permission(svcs.RBAC, domain.PermissionQueueDelete), handlers.Queue.Delete)
		queueGroup.POST("/:id/messages", httputil.Permission(svcs.RBAC, domain.PermissionQueueWrite), handlers.Queue.SendMessage)
		queueGroup.GET("/:id/messages", httputil.Permission(svcs.RBAC, domain.PermissionQueueRead), handlers.Queue.ReceiveMessages)
		queueGroup.DELETE("/:id/messages/:handle", httputil.Permission(svcs.RBAC, domain.PermissionQueueWrite), handlers.Queue.DeleteMessage)
		queueGroup.POST("/:id/purge", httputil.Permission(svcs.RBAC, domain.PermissionQueueWrite), handlers.Queue.Purge)
	}

	notifyGroup := r.Group("/notify")
	notifyGroup.Use(httputil.Auth(svcs.Identity))
	{
		notifyGroup.POST("/topics", httputil.Permission(svcs.RBAC, domain.PermissionNotifyCreate), handlers.Notify.CreateTopic)
		notifyGroup.GET("/topics", httputil.Permission(svcs.RBAC, domain.PermissionNotifyRead), handlers.Notify.ListTopics)
		notifyGroup.DELETE("/topics/:id", httputil.Permission(svcs.RBAC, domain.PermissionNotifyDelete), handlers.Notify.DeleteTopic)
		notifyGroup.POST("/topics/:id/subscriptions", httputil.Permission(svcs.RBAC, domain.PermissionNotifyWrite), handlers.Notify.Subscribe)
		notifyGroup.GET("/topics/:id/subscriptions", httputil.Permission(svcs.RBAC, domain.PermissionNotifyRead), handlers.Notify.ListSubscriptions)
		notifyGroup.DELETE("/subscriptions/:id", httputil.Permission(svcs.RBAC, domain.PermissionNotifyDelete), handlers.Notify.Unsubscribe)
		notifyGroup.POST("/topics/:id/publish", httputil.Permission(svcs.RBAC, domain.PermissionNotifyWrite), handlers.Notify.Publish)
	}

	cronGroup := r.Group("/cron")
	cronGroup.Use(httputil.Auth(svcs.Identity))
	{
		cronGroup.POST("/jobs", httputil.Permission(svcs.RBAC, domain.PermissionCronCreate), handlers.Cron.CreateJob)
		cronGroup.GET("/jobs", httputil.Permission(svcs.RBAC, domain.PermissionCronRead), handlers.Cron.ListJobs)
		cronGroup.GET("/jobs/:id", httputil.Permission(svcs.RBAC, domain.PermissionCronRead), handlers.Cron.GetJob)
		cronGroup.DELETE("/jobs/:id", httputil.Permission(svcs.RBAC, domain.PermissionCronDelete), handlers.Cron.DeleteJob)
		cronGroup.POST("/jobs/:id/pause", httputil.Permission(svcs.RBAC, domain.PermissionCronUpdate), handlers.Cron.PauseJob)
		cronGroup.POST("/jobs/:id/resume", httputil.Permission(svcs.RBAC, domain.PermissionCronUpdate), handlers.Cron.ResumeJob)
	}

	gatewayGroup := r.Group("/gateway")
	gatewayGroup.Use(httputil.Auth(svcs.Identity))
	{
		gatewayGroup.POST("/routes", httputil.Permission(svcs.RBAC, domain.PermissionGatewayCreate), handlers.Gateway.CreateRoute)
		gatewayGroup.GET("/routes", httputil.Permission(svcs.RBAC, domain.PermissionGatewayRead), handlers.Gateway.ListRoutes)
		gatewayGroup.DELETE("/routes/:id", httputil.Permission(svcs.RBAC, domain.PermissionGatewayDelete), handlers.Gateway.DeleteRoute)
	}

	containerGroup := r.Group("/containers")
	containerGroup.Use(httputil.Auth(svcs.Identity))
	{
		containerGroup.POST("/deployments", httputil.Permission(svcs.RBAC, domain.PermissionContainerCreate), handlers.Container.CreateDeployment)
		containerGroup.GET("/deployments", httputil.Permission(svcs.RBAC, domain.PermissionContainerRead), handlers.Container.ListDeployments)
		containerGroup.GET("/deployments/:id", httputil.Permission(svcs.RBAC, domain.PermissionContainerRead), handlers.Container.GetDeployment)
		containerGroup.POST("/deployments/:id/scale", httputil.Permission(svcs.RBAC, domain.PermissionContainerUpdate), handlers.Container.ScaleDeployment)
		containerGroup.DELETE("/deployments/:id", httputil.Permission(svcs.RBAC, domain.PermissionContainerDelete), handlers.Container.DeleteDeployment)
	}

	asgGroup := r.Group("/autoscaling")
	asgGroup.Use(httputil.Auth(svcs.Identity))
	{
		asgGroup.POST("/groups", httputil.Permission(svcs.RBAC, domain.PermissionAsCreate), handlers.AutoScaling.CreateGroup)
		asgGroup.GET("/groups", httputil.Permission(svcs.RBAC, domain.PermissionAsRead), handlers.AutoScaling.ListGroups)
		asgGroup.GET("/groups/:id", httputil.Permission(svcs.RBAC, domain.PermissionAsRead), handlers.AutoScaling.GetGroup)
		asgGroup.DELETE("/groups/:id", httputil.Permission(svcs.RBAC, domain.PermissionAsDelete), handlers.AutoScaling.DeleteGroup)
		asgGroup.POST("/groups/:id/policies", httputil.Permission(svcs.RBAC, domain.PermissionAsUpdate), handlers.AutoScaling.CreatePolicy)
		asgGroup.DELETE("/policies/:id", httputil.Permission(svcs.RBAC, domain.PermissionAsDelete), handlers.AutoScaling.DeletePolicy)
	}

	iacGroup := r.Group("/iac")
	iacGroup.Use(httputil.Auth(svcs.Identity))
	{
		iacGroup.POST("/stacks", httputil.Permission(svcs.RBAC, domain.PermissionStackCreate), handlers.Stack.Create)
		iacGroup.GET("/stacks", httputil.Permission(svcs.RBAC, domain.PermissionStackRead), handlers.Stack.List)
		iacGroup.GET("/stacks/:id", httputil.Permission(svcs.RBAC, domain.PermissionStackRead), handlers.Stack.Get)
		iacGroup.DELETE("/stacks/:id", httputil.Permission(svcs.RBAC, domain.PermissionStackDelete), handlers.Stack.Delete)
		iacGroup.POST("/validate", httputil.Permission(svcs.RBAC, domain.PermissionStackRead), handlers.Stack.Validate)
	}
}

func registerAdminRoutes(r *gin.Engine, handlers *Handlers, svcs *Services) {
	eventGroup := r.Group("/events")
	eventGroup.Use(httputil.Auth(svcs.Identity))
	{
		eventGroup.GET("", handlers.Event.List)
	}

	auditGroup := r.Group("/audit")
	auditGroup.Use(httputil.Auth(svcs.Identity))
	{
		auditGroup.GET("", handlers.Audit.ListLogs)
	}

	rbacGroup := r.Group("/rbac")
	rbacGroup.Use(httputil.Auth(svcs.Identity))
	{
		rbacGroup.POST("/roles", httputil.Permission(svcs.RBAC, domain.PermissionFullAccess), handlers.RBAC.CreateRole)
		rbacGroup.GET("/roles", httputil.Permission(svcs.RBAC, domain.PermissionFullAccess), handlers.RBAC.ListRoles)
		rbacGroup.GET(roleIDRoute, httputil.Permission(svcs.RBAC, domain.PermissionFullAccess), handlers.RBAC.GetRole)
		rbacGroup.PUT(roleIDRoute, httputil.Permission(svcs.RBAC, domain.PermissionFullAccess), handlers.RBAC.UpdateRole)
		rbacGroup.DELETE(roleIDRoute, httputil.Permission(svcs.RBAC, domain.PermissionFullAccess), handlers.RBAC.DeleteRole)
		rbacGroup.POST("/roles/:id/permissions", httputil.Permission(svcs.RBAC, domain.PermissionFullAccess), handlers.RBAC.AddPermission)
		rbacGroup.DELETE("/roles/:id/permissions/:permission", httputil.Permission(svcs.RBAC, domain.PermissionFullAccess), handlers.RBAC.RemovePermission)
		rbacGroup.POST("/bindings", httputil.Permission(svcs.RBAC, domain.PermissionFullAccess), handlers.RBAC.BindRole)
		rbacGroup.GET("/bindings", httputil.Permission(svcs.RBAC, domain.PermissionFullAccess), handlers.RBAC.ListRoleBindings)
	}

	dashboardGroup := r.Group("/api/dashboard")
	dashboardGroup.Use(httputil.Auth(svcs.Identity))
	{
		dashboardGroup.GET("/summary", handlers.Dashboard.GetSummary)
		dashboardGroup.GET("/events", handlers.Dashboard.GetRecentEvents)
		dashboardGroup.GET("/stats", handlers.Dashboard.GetStats)
		dashboardGroup.GET("/stream", handlers.Dashboard.StreamEvents)
		dashboardGroup.GET("/ws", handlers.Ws.ServeWS)
	}

	billingGroup := r.Group("/billing")
	billingGroup.Use(httputil.Auth(svcs.Identity))
	{
		billingGroup.GET("/summary", handlers.Accounting.GetSummary)
		billingGroup.GET("/usage", handlers.Accounting.ListUsage)
	}
}
