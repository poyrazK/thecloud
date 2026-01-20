// Package setup wires API dependencies and routes.
package setup

import (
	"fmt"
	"log/slog"

	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/handlers/ws"
	"github.com/poyrazk/thecloud/internal/platform"
	"github.com/poyrazk/thecloud/internal/repositories/filesystem"
	"github.com/poyrazk/thecloud/internal/repositories/k8s"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/poyrazk/thecloud/internal/repositories/redis"
	"github.com/poyrazk/thecloud/internal/workers"
	redisv9 "github.com/redis/go-redis/v9"
)

// Repositories bundles all data access implementations.
type Repositories struct {
	Audit         ports.AuditRepository
	User          ports.UserRepository
	Identity      ports.IdentityRepository
	PasswordReset ports.PasswordResetRepository
	RBAC          ports.RoleRepository
	Instance      ports.InstanceRepository
	Vpc           ports.VpcRepository
	Event         ports.EventRepository
	Volume        ports.VolumeRepository
	SecurityGroup ports.SecurityGroupRepository
	Subnet        ports.SubnetRepository
	LB            ports.LBRepository
	Snapshot      ports.SnapshotRepository
	Stack         ports.StackRepository
	Storage       ports.StorageRepository
	Database      ports.DatabaseRepository
	Secret        ports.SecretRepository
	Function      ports.FunctionRepository
	Cache         ports.CacheRepository
	Queue         ports.QueueRepository
	Notify        ports.NotifyRepository
	Cron          ports.CronRepository
	Gateway       ports.GatewayRepository
	Container     ports.ContainerRepository
	AutoScaling   ports.AutoScalingRepository
	Accounting    ports.AccountingRepository
	TaskQueue     ports.TaskQueue
	Image         ports.ImageRepository
	Cluster       ports.ClusterRepository
}

// InitRepositories constructs repositories using the provided database clients.
func InitRepositories(db postgres.DB, rdb *redisv9.Client) *Repositories {
	return &Repositories{
		Audit:         postgres.NewAuditRepository(db),
		User:          postgres.NewUserRepo(db),
		Identity:      postgres.NewIdentityRepository(db),
		PasswordReset: postgres.NewPasswordResetRepository(db),
		RBAC:          postgres.NewRBACRepository(db),
		Instance:      postgres.NewInstanceRepository(db),
		Vpc:           postgres.NewVpcRepository(db),
		Event:         postgres.NewEventRepository(db),
		Volume:        postgres.NewVolumeRepository(db),
		SecurityGroup: postgres.NewSecurityGroupRepository(db),
		Subnet:        postgres.NewSubnetRepository(db),
		LB:            postgres.NewLBRepository(db),
		Snapshot:      postgres.NewSnapshotRepository(db),
		Stack:         postgres.NewStackRepository(db),
		Storage:       postgres.NewStorageRepository(db),
		Database:      postgres.NewDatabaseRepository(db),
		Secret:        postgres.NewSecretRepository(db),
		Function:      postgres.NewFunctionRepository(db),
		Cache:         postgres.NewCacheRepository(db),
		Queue:         postgres.NewPostgresQueueRepository(db),
		Notify:        postgres.NewPostgresNotifyRepository(db),
		Cron:          postgres.NewPostgresCronRepository(db),
		Gateway:       postgres.NewPostgresGatewayRepository(db),
		Container:     postgres.NewPostgresContainerRepository(db),
		AutoScaling:   postgres.NewAutoScalingRepo(db),
		Accounting:    postgres.NewAccountingRepository(db),
		TaskQueue:     redis.NewRedisTaskQueue(rdb),
		Image:         postgres.NewImageRepository(db),
		Cluster:       postgres.NewClusterRepository(db),
	}
}

// Services bundles the core application services.
type Services struct {
	WsHub         *ws.Hub
	Audit         ports.AuditService
	Identity      ports.IdentityService
	Auth          ports.AuthService
	PasswordReset ports.PasswordResetService
	RBAC          ports.RBACService
	Vpc           ports.VpcService
	Subnet        ports.SubnetService
	Event         ports.EventService
	Volume        ports.VolumeService
	Instance      ports.InstanceService
	SecurityGroup ports.SecurityGroupService
	LB            ports.LBService
	Dashboard     ports.DashboardService
	Snapshot      ports.SnapshotService
	Stack         ports.StackService
	Storage       ports.StorageService
	Database      ports.DatabaseService
	Secret        ports.SecretService
	Function      ports.FunctionService
	Cache         ports.CacheService
	Queue         ports.QueueService
	Notify        ports.NotifyService
	Cron          ports.CronService
	Gateway       ports.GatewayService
	Container     ports.ContainerService
	Health        ports.HealthService
	AutoScaling   ports.AutoScalingService
	Accounting    ports.AccountingService
	Image         ports.ImageService
	Cluster       ports.ClusterService
}

// Workers struct to return background workers
type Workers struct {
	LB          *services.LBWorker
	AutoScaling *services.AutoScalingWorker
	Cron        *services.CronWorker
	Container   *services.ContainerWorker
	Provision   *workers.ProvisionWorker
	Accounting  *workers.AccountingWorker
	Cluster     *workers.ClusterWorker
}

// ServiceConfig holds the dependencies required to initialize services
type ServiceConfig struct {
	Config  *platform.Config
	Repos   *Repositories
	Compute ports.ComputeBackend
	Storage ports.StorageBackend
	Network ports.NetworkBackend
	LBProxy ports.LBProxyAdapter
	DB      postgres.DB
	RDB     *redisv9.Client
	Logger  *slog.Logger
}

// InitServices constructs core services and background workers.
func InitServices(c ServiceConfig) (*Services, *Workers, error) {
	// 1. Core Services (Audit, Identity, Auth, RBAC)
	auditSvc := services.NewAuditService(c.Repos.Audit)
	identitySvc := initIdentityServices(c, auditSvc)
	authSvc := services.NewAuthService(c.Repos.User, identitySvc, auditSvc)
	pwdResetSvc := services.NewPasswordResetService(c.Repos.PasswordReset, c.Repos.User, c.Logger)
	rbacSvc := initRBACServices(c)

	// 2. WebSocket & Core Infrastructure
	wsHub := ws.NewHub(c.Logger)
	go wsHub.Run()
	eventSvc := services.NewEventService(c.Repos.Event, wsHub, c.Logger)

	// 3. Cloud Infrastructure Services (VPC, Subnet, Instance, Volume, SG, LB)
	vpcSvc := services.NewVpcService(c.Repos.Vpc, c.Network, auditSvc, c.Logger, c.Config.DefaultVPCCIDR)
	subnetSvc := services.NewSubnetService(c.Repos.Subnet, c.Repos.Vpc, auditSvc, c.Logger)
	volumeSvc := services.NewVolumeService(c.Repos.Volume, c.Storage, eventSvc, auditSvc, c.Logger)
	instSvcConcrete := services.NewInstanceService(services.InstanceServiceParams{
		Repo: c.Repos.Instance, VpcRepo: c.Repos.Vpc, SubnetRepo: c.Repos.Subnet, VolumeRepo: c.Repos.Volume,
		Compute: c.Compute, Network: c.Network, EventSvc: eventSvc, AuditSvc: auditSvc, TaskQueue: c.Repos.TaskQueue, Logger: c.Logger,
	})
	sgSvc := services.NewSecurityGroupService(c.Repos.SecurityGroup, c.Repos.Vpc, c.Network, auditSvc, c.Logger)

	lbSvc := services.NewLBService(c.Repos.LB, c.Repos.Vpc, c.Repos.Instance, auditSvc)
	lbWorker := services.NewLBWorker(c.Repos.LB, c.Repos.Instance, c.LBProxy)

	// 4. Advanced Services (Storage, DB, Secrets, FaaS, Cache, Queue)
	fileStore, err := filesystem.NewLocalFileStore("./thecloud-data/local/storage")
	if err != nil {
		return nil, nil, err
	}
	storageSvc := services.NewStorageService(c.Repos.Storage, fileStore, auditSvc)

	databaseSvc := services.NewDatabaseService(c.Repos.Database, c.Compute, c.Repos.Vpc, eventSvc, auditSvc, c.Logger)
	secretSvc := services.NewSecretService(c.Repos.Secret, eventSvc, auditSvc, c.Logger, c.Config.SecretsEncryptionKey, c.Config.Environment)
	fnSvc := services.NewFunctionService(c.Repos.Function, c.Compute, fileStore, auditSvc, c.Logger)
	cacheSvc := services.NewCacheService(c.Repos.Cache, c.Compute, c.Repos.Vpc, eventSvc, auditSvc, c.Logger)
	queueSvc := services.NewQueueService(c.Repos.Queue, eventSvc, auditSvc)
	notifySvc := services.NewNotifyService(c.Repos.Notify, queueSvc, eventSvc, auditSvc, c.Logger)

	// 5. DevOps & Automation Services
	cronSvc := services.NewCronService(c.Repos.Cron, eventSvc, auditSvc)
	cronWorker := services.NewCronWorker(c.Repos.Cron)
	gwSvc := services.NewGatewayService(c.Repos.Gateway, auditSvc)
	containerSvc := services.NewContainerService(c.Repos.Container, eventSvc, auditSvc)
	containerWorker := services.NewContainerWorker(c.Repos.Container, instSvcConcrete, eventSvc)
	snapshotSvc := services.NewSnapshotService(c.Repos.Snapshot, c.Repos.Volume, c.Storage, eventSvc, auditSvc, c.Logger)
	stackSvc := services.NewStackService(c.Repos.Stack, instSvcConcrete, vpcSvc, volumeSvc, snapshotSvc, c.Logger)

	// 6. Business & Scaling Services
	asgSvc := services.NewAutoScalingService(c.Repos.AutoScaling, c.Repos.Vpc, auditSvc)
	asgWorker := services.NewAutoScalingWorker(c.Repos.AutoScaling, instSvcConcrete, lbSvc, eventSvc, ports.RealClock{})
	accountingSvc := services.NewAccountingService(c.Repos.Accounting, c.Repos.Instance, c.Logger)
	accountingWorker := workers.NewAccountingWorker(accountingSvc, c.Logger)
	imageSvc := services.NewImageService(c.Repos.Image, fileStore, c.Logger)
	provisionWorker := workers.NewProvisionWorker(instSvcConcrete, c.Repos.TaskQueue, c.Logger)

	clusterProvisioner := k8s.NewKubeadmProvisioner(instSvcConcrete, c.Repos.Cluster, secretSvc, sgSvc, storageSvc, lbSvc, c.Logger)
	clusterSvc, err := services.NewClusterService(services.ClusterServiceParams{
		Repo: c.Repos.Cluster, Provisioner: clusterProvisioner, VpcSvc: vpcSvc, InstanceSvc: instSvcConcrete, SecretSvc: secretSvc, TaskQueue: c.Repos.TaskQueue, Logger: c.Logger,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to init cluster service: %w", err)
	}

	svcs := &Services{
		WsHub: wsHub, Audit: auditSvc, Identity: identitySvc, Auth: authSvc, PasswordReset: pwdResetSvc, RBAC: rbacSvc,
		Vpc: vpcSvc, Subnet: subnetSvc, Event: eventSvc, Volume: volumeSvc, Instance: instSvcConcrete,
		SecurityGroup: sgSvc, LB: lbSvc, Snapshot: snapshotSvc, Stack: stackSvc,
		Storage: storageSvc, Database: databaseSvc, Secret: secretSvc, Function: fnSvc, Cache: cacheSvc,
		Queue: queueSvc, Notify: notifySvc, Cron: cronSvc, Gateway: gwSvc, Container: containerSvc,
		Health: services.NewHealthServiceImpl(c.DB, c.Compute), AutoScaling: asgSvc, Accounting: accountingSvc, Image: imageSvc,
		Cluster:   clusterSvc,
		Dashboard: services.NewDashboardService(c.Repos.Instance, c.Repos.Volume, c.Repos.Vpc, c.Repos.Event, c.Logger),
	}

	workersCollection := &Workers{
		LB: lbWorker, AutoScaling: asgWorker, Cron: cronWorker, Container: containerWorker,
		Provision: provisionWorker, Accounting: accountingWorker,
		Cluster: workers.NewClusterWorker(c.Repos.Cluster, clusterProvisioner, c.Repos.TaskQueue, c.Logger),
	}

	return svcs, workersCollection, nil
}

func initIdentityServices(c ServiceConfig, audit ports.AuditService) ports.IdentityService {
	base := services.NewIdentityService(c.Repos.Identity, audit)
	return services.NewCachedIdentityService(base, c.RDB, c.Logger)
}

func initRBACServices(c ServiceConfig) ports.RBACService {
	base := services.NewRBACService(c.Repos.User, c.Repos.RBAC, c.Logger)
	return services.NewCachedRBACService(base, c.RDB, c.Logger)
}
