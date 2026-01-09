package setup

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/poyrazk/thecloud/docs/swagger"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	httphandlers "github.com/poyrazk/thecloud/internal/handlers"
	"github.com/poyrazk/thecloud/internal/platform"
	"github.com/poyrazk/thecloud/pkg/httputil"
	"github.com/poyrazk/thecloud/pkg/ratelimit"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/time/rate"
)

const (
	bucketKeyRoute = "/:bucket/:key"
	roleIDRoute    = "/roles/:id"
)

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
}

func InitHandlers(svcs *Services) *Handlers {
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
		Storage:       httphandlers.NewStorageHandler(svcs.Storage),
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
	}
}

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

	// Rate Limiter (5 req/sec, burst 10)
	limiter := ratelimit.NewIPRateLimiter(rate.Limit(5), 10, logger)
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

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Auth Rate Limiter (5 req/min, burst 5)
	authLimiter := ratelimit.NewIPRateLimiter(rate.Limit(5.0/60.0), 5, logger)
	authMiddleware := ratelimit.Middleware(authLimiter)

	// Identity Routes
	r.POST("/auth/register", authMiddleware, handlers.Auth.Register)
	r.POST("/auth/login", authMiddleware, handlers.Auth.Login)
	r.POST("/auth/forgot-password", authMiddleware, handlers.Auth.ForgotPassword)
	r.POST("/auth/reset-password", authMiddleware, handlers.Auth.ResetPassword)

	keyGroup := r.Group("/auth/keys")
	keyGroup.Use(httputil.Auth(services.Identity))
	{
		keyGroup.POST("", handlers.Identity.CreateKey)
		keyGroup.GET("", handlers.Identity.ListKeys)
		keyGroup.DELETE("/:id", handlers.Identity.RevokeKey)
		keyGroup.POST("/:id/rotate", handlers.Identity.RotateKey)
		keyGroup.POST("/:id/regenerate", handlers.Identity.RegenerateKey)
	}

	// Instance Routes (Protected)
	instanceGroup := r.Group("/instances")
	instanceGroup.Use(httputil.Auth(services.Identity))
	{
		instanceGroup.POST("", httputil.Permission(services.RBAC, domain.PermissionInstanceLaunch), handlers.Instance.Launch)
		instanceGroup.GET("", httputil.Permission(services.RBAC, domain.PermissionInstanceRead), handlers.Instance.List)
		instanceGroup.GET("/:id", httputil.Permission(services.RBAC, domain.PermissionInstanceRead), handlers.Instance.Get)
		instanceGroup.POST("/:id/stop", httputil.Permission(services.RBAC, domain.PermissionInstanceUpdate), handlers.Instance.Stop)
		instanceGroup.GET("/:id/logs", httputil.Permission(services.RBAC, domain.PermissionInstanceRead), handlers.Instance.GetLogs)
		instanceGroup.GET("/:id/stats", httputil.Permission(services.RBAC, domain.PermissionInstanceRead), handlers.Instance.GetStats)
		instanceGroup.DELETE("/:id", httputil.Permission(services.RBAC, domain.PermissionInstanceTerminate), handlers.Instance.Terminate)
	}

	// VPC Routes (Protected)
	vpcGroup := r.Group("/vpcs")
	vpcGroup.Use(httputil.Auth(services.Identity))
	{
		vpcGroup.POST("", httputil.Permission(services.RBAC, domain.PermissionVpcCreate), handlers.Vpc.Create)
		vpcGroup.GET("", httputil.Permission(services.RBAC, domain.PermissionVpcRead), handlers.Vpc.List)
		vpcGroup.GET("/:id", httputil.Permission(services.RBAC, domain.PermissionVpcRead), handlers.Vpc.Get)
		vpcGroup.DELETE("/:id", httputil.Permission(services.RBAC, domain.PermissionVpcDelete), handlers.Vpc.Delete)

		// Subnet routes nested under VPC
		vpcGroup.POST("/:vpc_id/subnets", httputil.Permission(services.RBAC, domain.PermissionVpcUpdate), handlers.Subnet.Create)
		vpcGroup.GET("/:vpc_id/subnets", httputil.Permission(services.RBAC, domain.PermissionVpcRead), handlers.Subnet.List)
	}

	// standalone Subnet routes
	subnetGroup := r.Group("/subnets")
	subnetGroup.Use(httputil.Auth(services.Identity))
	{
		subnetGroup.GET("/:id", httputil.Permission(services.RBAC, domain.PermissionVpcRead), handlers.Subnet.Get)
		subnetGroup.DELETE("/:id", httputil.Permission(services.RBAC, domain.PermissionVpcUpdate), handlers.Subnet.Delete)
	}

	// Security Group Routes (Protected)
	sgGroup := r.Group("/security-groups")
	sgGroup.Use(httputil.Auth(services.Identity))
	{
		sgGroup.POST("", handlers.SecurityGroup.Create)
		sgGroup.GET("", handlers.SecurityGroup.List)
		sgGroup.GET("/:id", handlers.SecurityGroup.Get)
		sgGroup.DELETE("/:id", handlers.SecurityGroup.Delete)
		sgGroup.POST("/:id/rules", handlers.SecurityGroup.AddRule)
		sgGroup.POST("/attach", handlers.SecurityGroup.Attach)
	}

	// Storage Routes (Protected)
	storageGroup := r.Group("/storage")
	storageGroup.Use(httputil.Auth(services.Identity))
	{
		storageGroup.PUT(bucketKeyRoute, handlers.Storage.Upload)
		storageGroup.GET(bucketKeyRoute, handlers.Storage.Download)
		storageGroup.GET("/:bucket", handlers.Storage.List)
		storageGroup.DELETE(bucketKeyRoute, handlers.Storage.Delete)
	}

	// Event Routes (Protected)
	eventGroup := r.Group("/events")
	eventGroup.Use(httputil.Auth(services.Identity))
	{
		eventGroup.GET("", handlers.Event.List)
	}

	// Audit Routes (Protected)
	auditGroup := r.Group("/audit")
	auditGroup.Use(httputil.Auth(services.Identity))
	{
		auditGroup.GET("", handlers.Audit.ListLogs)
	}

	// RBAC Routes (Protected)
	rbacGroup := r.Group("/rbac")
	rbacGroup.Use(httputil.Auth(services.Identity))
	{
		rbacGroup.POST("/roles", httputil.Permission(services.RBAC, domain.PermissionFullAccess), handlers.RBAC.CreateRole)
		rbacGroup.GET("/roles", httputil.Permission(services.RBAC, domain.PermissionFullAccess), handlers.RBAC.ListRoles)
		rbacGroup.GET(roleIDRoute, httputil.Permission(services.RBAC, domain.PermissionFullAccess), handlers.RBAC.GetRole)
		rbacGroup.PUT(roleIDRoute, httputil.Permission(services.RBAC, domain.PermissionFullAccess), handlers.RBAC.UpdateRole)
		rbacGroup.DELETE(roleIDRoute, httputil.Permission(services.RBAC, domain.PermissionFullAccess), handlers.RBAC.DeleteRole)
		rbacGroup.POST("/roles/:id/permissions", httputil.Permission(services.RBAC, domain.PermissionFullAccess), handlers.RBAC.AddPermission)
		rbacGroup.DELETE("/roles/:id/permissions/:permission", httputil.Permission(services.RBAC, domain.PermissionFullAccess), handlers.RBAC.RemovePermission)
		rbacGroup.POST("/bindings", httputil.Permission(services.RBAC, domain.PermissionFullAccess), handlers.RBAC.BindRole)
		rbacGroup.GET("/bindings", httputil.Permission(services.RBAC, domain.PermissionFullAccess), handlers.RBAC.ListRoleBindings)
	}

	// Volume Routes (Protected)
	volumeGroup := r.Group("/volumes")
	volumeGroup.Use(httputil.Auth(services.Identity))
	{
		volumeGroup.POST("", httputil.Permission(services.RBAC, domain.PermissionVolumeCreate), handlers.Volume.Create)
		volumeGroup.GET("", httputil.Permission(services.RBAC, domain.PermissionVolumeRead), handlers.Volume.List)
		volumeGroup.GET("/:id", httputil.Permission(services.RBAC, domain.PermissionVolumeRead), handlers.Volume.Get)
		volumeGroup.DELETE("/:id", httputil.Permission(services.RBAC, domain.PermissionVolumeDelete), handlers.Volume.Delete)
	}

	// Dashboard Routes (Protected)
	dashboardGroup := r.Group("/api/dashboard")
	dashboardGroup.Use(httputil.Auth(services.Identity))
	{
		dashboardGroup.GET("/summary", handlers.Dashboard.GetSummary)
		dashboardGroup.GET("/events", handlers.Dashboard.GetRecentEvents)
		dashboardGroup.GET("/stats", handlers.Dashboard.GetStats)
		dashboardGroup.GET("/stream", handlers.Dashboard.StreamEvents)
	}

	// Snapshot Routes (Protected)
	snapshotGroup := r.Group("/snapshots")
	snapshotGroup.Use(httputil.Auth(services.Identity))
	{
		snapshotGroup.POST("", httputil.Permission(services.RBAC, domain.PermissionSnapshotCreate), handlers.Snapshot.Create)
		snapshotGroup.GET("", httputil.Permission(services.RBAC, domain.PermissionSnapshotRead), handlers.Snapshot.List)
		snapshotGroup.GET("/:id", httputil.Permission(services.RBAC, domain.PermissionSnapshotRead), handlers.Snapshot.Get)
		snapshotGroup.DELETE("/:id", httputil.Permission(services.RBAC, domain.PermissionSnapshotDelete), handlers.Snapshot.Delete)
		snapshotGroup.POST("/:id/restore", httputil.Permission(services.RBAC, domain.PermissionSnapshotRestore), handlers.Snapshot.Restore)
	}

	// Load Balancer Routes (Protected)
	lbGroup := r.Group("/lb")
	lbGroup.Use(httputil.Auth(services.Identity))
	{
		lbGroup.POST("", httputil.Permission(services.RBAC, domain.PermissionLbCreate), handlers.LB.Create)
		lbGroup.GET("", httputil.Permission(services.RBAC, domain.PermissionLbRead), handlers.LB.List)
		lbGroup.GET("/:id", httputil.Permission(services.RBAC, domain.PermissionLbRead), handlers.LB.Get)
		lbGroup.DELETE("/:id", httputil.Permission(services.RBAC, domain.PermissionLbDelete), handlers.LB.Delete)
		lbGroup.POST("/:id/targets", httputil.Permission(services.RBAC, domain.PermissionLbUpdate), handlers.LB.AddTarget)
		lbGroup.GET("/:id/targets", httputil.Permission(services.RBAC, domain.PermissionLbRead), handlers.LB.ListTargets)
		lbGroup.DELETE("/:id/targets/:instanceId", httputil.Permission(services.RBAC, domain.PermissionLbUpdate), handlers.LB.RemoveTarget)
	}

	// Database Routes (Protected)
	dbGroup := r.Group("/databases")
	dbGroup.Use(httputil.Auth(services.Identity))
	{
		dbGroup.POST("", httputil.Permission(services.RBAC, domain.PermissionDBCreate), handlers.Database.Create)
		dbGroup.GET("", httputil.Permission(services.RBAC, domain.PermissionDBRead), handlers.Database.List)
		dbGroup.GET("/:id", httputil.Permission(services.RBAC, domain.PermissionDBRead), handlers.Database.Get)
		dbGroup.DELETE("/:id", httputil.Permission(services.RBAC, domain.PermissionDBDelete), handlers.Database.Delete)
		dbGroup.GET("/:id/connection", httputil.Permission(services.RBAC, domain.PermissionDBRead), handlers.Database.GetConnectionString)
	}

	// Secret Routes (Protected)
	secretGroup := r.Group("/secrets")
	secretGroup.Use(httputil.Auth(services.Identity))
	{
		secretGroup.POST("", httputil.Permission(services.RBAC, domain.PermissionSecretCreate), handlers.Secret.Create)
		secretGroup.GET("", httputil.Permission(services.RBAC, domain.PermissionSecretRead), handlers.Secret.List)
		secretGroup.GET("/:id", httputil.Permission(services.RBAC, domain.PermissionSecretRead), handlers.Secret.Get)
		secretGroup.DELETE("/:id", httputil.Permission(services.RBAC, domain.PermissionSecretDelete), handlers.Secret.Delete)
	}

	// Function Routes (Protected)
	fnGroup := r.Group("/functions")
	fnGroup.Use(httputil.Auth(services.Identity))
	{
		fnGroup.POST("", httputil.Permission(services.RBAC, domain.PermissionFunctionCreate), handlers.Function.Create)
		fnGroup.GET("", httputil.Permission(services.RBAC, domain.PermissionFunctionRead), handlers.Function.List)
		fnGroup.GET("/:id", httputil.Permission(services.RBAC, domain.PermissionFunctionRead), handlers.Function.Get)
		fnGroup.DELETE("/:id", httputil.Permission(services.RBAC, domain.PermissionFunctionDelete), handlers.Function.Delete)
		fnGroup.POST("/:id/invoke", httputil.Permission(services.RBAC, domain.PermissionFunctionInvoke), handlers.Function.Invoke)
		fnGroup.GET("/:id/logs", httputil.Permission(services.RBAC, domain.PermissionFunctionRead), handlers.Function.GetLogs)
	}

	// Cache Routes (Protected)
	cacheGroup := r.Group("/caches")
	cacheGroup.Use(httputil.Auth(services.Identity))
	{
		cacheGroup.POST("", httputil.Permission(services.RBAC, domain.PermissionCacheCreate), handlers.Cache.Create)
		cacheGroup.GET("", httputil.Permission(services.RBAC, domain.PermissionCacheRead), handlers.Cache.List)
		cacheGroup.GET("/:id", httputil.Permission(services.RBAC, domain.PermissionCacheRead), handlers.Cache.Get)
		cacheGroup.DELETE("/:id", httputil.Permission(services.RBAC, domain.PermissionCacheDelete), handlers.Cache.Delete)
		cacheGroup.GET("/:id/connection", httputil.Permission(services.RBAC, domain.PermissionCacheRead), handlers.Cache.GetConnectionString)
		cacheGroup.POST("/:id/flush", httputil.Permission(services.RBAC, domain.PermissionCacheUpdate), handlers.Cache.Flush)
		cacheGroup.GET("/:id/stats", httputil.Permission(services.RBAC, domain.PermissionCacheRead), handlers.Cache.GetStats)
	}

	queueGroup := r.Group("/queues")
	queueGroup.Use(httputil.Auth(services.Identity))
	{
		queueGroup.POST("", httputil.Permission(services.RBAC, domain.PermissionQueueCreate), handlers.Queue.Create)
		queueGroup.GET("", httputil.Permission(services.RBAC, domain.PermissionQueueRead), handlers.Queue.List)
		queueGroup.GET("/:id", httputil.Permission(services.RBAC, domain.PermissionQueueRead), handlers.Queue.Get)
		queueGroup.DELETE("/:id", httputil.Permission(services.RBAC, domain.PermissionQueueDelete), handlers.Queue.Delete)
		queueGroup.POST("/:id/messages", httputil.Permission(services.RBAC, domain.PermissionQueueWrite), handlers.Queue.SendMessage)
		queueGroup.GET("/:id/messages", httputil.Permission(services.RBAC, domain.PermissionQueueRead), handlers.Queue.ReceiveMessages)
		queueGroup.DELETE("/:id/messages/:handle", httputil.Permission(services.RBAC, domain.PermissionQueueWrite), handlers.Queue.DeleteMessage)
		queueGroup.POST("/:id/purge", httputil.Permission(services.RBAC, domain.PermissionQueueWrite), handlers.Queue.Purge)
	}

	notifyGroup := r.Group("/notify")
	notifyGroup.Use(httputil.Auth(services.Identity))
	{
		notifyGroup.POST("/topics", httputil.Permission(services.RBAC, domain.PermissionNotifyCreate), handlers.Notify.CreateTopic)
		notifyGroup.GET("/topics", httputil.Permission(services.RBAC, domain.PermissionNotifyRead), handlers.Notify.ListTopics)
		notifyGroup.DELETE("/topics/:id", httputil.Permission(services.RBAC, domain.PermissionNotifyDelete), handlers.Notify.DeleteTopic)
		notifyGroup.POST("/topics/:id/subscriptions", httputil.Permission(services.RBAC, domain.PermissionNotifyWrite), handlers.Notify.Subscribe)
		notifyGroup.GET("/topics/:id/subscriptions", httputil.Permission(services.RBAC, domain.PermissionNotifyRead), handlers.Notify.ListSubscriptions)
		notifyGroup.DELETE("/subscriptions/:id", httputil.Permission(services.RBAC, domain.PermissionNotifyDelete), handlers.Notify.Unsubscribe)
		notifyGroup.POST("/topics/:id/publish", httputil.Permission(services.RBAC, domain.PermissionNotifyWrite), handlers.Notify.Publish)
	}

	cronGroup := r.Group("/cron")
	cronGroup.Use(httputil.Auth(services.Identity))
	{
		cronGroup.POST("/jobs", httputil.Permission(services.RBAC, domain.PermissionCronCreate), handlers.Cron.CreateJob)
		cronGroup.GET("/jobs", httputil.Permission(services.RBAC, domain.PermissionCronRead), handlers.Cron.ListJobs)
		cronGroup.GET("/jobs/:id", httputil.Permission(services.RBAC, domain.PermissionCronRead), handlers.Cron.GetJob)
		cronGroup.DELETE("/jobs/:id", httputil.Permission(services.RBAC, domain.PermissionCronDelete), handlers.Cron.DeleteJob)
		cronGroup.POST("/jobs/:id/pause", httputil.Permission(services.RBAC, domain.PermissionCronUpdate), handlers.Cron.PauseJob)
		cronGroup.POST("/jobs/:id/resume", httputil.Permission(services.RBAC, domain.PermissionCronUpdate), handlers.Cron.ResumeJob)
	}

	gatewayGroup := r.Group("/gateway")
	gatewayGroup.Use(httputil.Auth(services.Identity))
	{
		gatewayGroup.POST("/routes", httputil.Permission(services.RBAC, domain.PermissionGatewayCreate), handlers.Gateway.CreateRoute)
		gatewayGroup.GET("/routes", httputil.Permission(services.RBAC, domain.PermissionGatewayRead), handlers.Gateway.ListRoutes)
		gatewayGroup.DELETE("/routes/:id", httputil.Permission(services.RBAC, domain.PermissionGatewayDelete), handlers.Gateway.DeleteRoute)
	}

	// The actual Gateway Proxy (Public)
	r.Any("/gw/*proxy", handlers.Gateway.Proxy)

	containerGroup := r.Group("/containers")
	containerGroup.Use(httputil.Auth(services.Identity))
	{
		containerGroup.POST("/deployments", httputil.Permission(services.RBAC, domain.PermissionContainerCreate), handlers.Container.CreateDeployment)
		containerGroup.GET("/deployments", httputil.Permission(services.RBAC, domain.PermissionContainerRead), handlers.Container.ListDeployments)
		containerGroup.GET("/deployments/:id", httputil.Permission(services.RBAC, domain.PermissionContainerRead), handlers.Container.GetDeployment)
		containerGroup.POST("/deployments/:id/scale", httputil.Permission(services.RBAC, domain.PermissionContainerUpdate), handlers.Container.ScaleDeployment)
		containerGroup.DELETE("/deployments/:id", httputil.Permission(services.RBAC, domain.PermissionContainerDelete), handlers.Container.DeleteDeployment)
	}

	// Auto-Scaling Routes (Protected)
	asgGroup := r.Group("/autoscaling")
	asgGroup.Use(httputil.Auth(services.Identity))
	{
		asgGroup.POST("/groups", httputil.Permission(services.RBAC, domain.PermissionAsCreate), handlers.AutoScaling.CreateGroup)
		asgGroup.GET("/groups", httputil.Permission(services.RBAC, domain.PermissionAsRead), handlers.AutoScaling.ListGroups)
		asgGroup.GET("/groups/:id", httputil.Permission(services.RBAC, domain.PermissionAsRead), handlers.AutoScaling.GetGroup)
		asgGroup.DELETE("/groups/:id", httputil.Permission(services.RBAC, domain.PermissionAsDelete), handlers.AutoScaling.DeleteGroup)
		asgGroup.POST("/groups/:id/policies", httputil.Permission(services.RBAC, domain.PermissionAsUpdate), handlers.AutoScaling.CreatePolicy)
		asgGroup.DELETE("/policies/:id", httputil.Permission(services.RBAC, domain.PermissionAsDelete), handlers.AutoScaling.DeletePolicy)
	}

	// IaC Routes (Protected)
	iacGroup := r.Group("/iac")
	iacGroup.Use(httputil.Auth(services.Identity))
	{
		iacGroup.POST("/stacks", httputil.Permission(services.RBAC, domain.PermissionStackCreate), handlers.Stack.Create)
		iacGroup.GET("/stacks", httputil.Permission(services.RBAC, domain.PermissionStackRead), handlers.Stack.List)
		iacGroup.GET("/stacks/:id", httputil.Permission(services.RBAC, domain.PermissionStackRead), handlers.Stack.Get)
		iacGroup.DELETE("/stacks/:id", httputil.Permission(services.RBAC, domain.PermissionStackDelete), handlers.Stack.Delete)
		iacGroup.POST("/validate", httputil.Permission(services.RBAC, domain.PermissionStackRead), handlers.Stack.Validate)
	}

	return r
}
