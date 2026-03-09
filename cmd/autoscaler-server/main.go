package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/poyrazk/thecloud/internal/autoscaler"
	"github.com/poyrazk/thecloud/internal/autoscaler/protos"
	"github.com/poyrazk/thecloud/pkg/sdk"
	"google.golang.org/grpc"
	"k8s.io/klog/v2"
)

const (
	defaultGRPCPort = 50051
	defaultAPIPort  = 443
)

func main() {
	var (
		port      = flag.Int("port", defaultGRPCPort, "The gRPC server port")
		apiURL    = flag.String("api-url", os.Getenv("CLOUD_API_URL"), "The Cloud API URL")
		apiKey    = flag.String("api-key", os.Getenv("CLOUD_API_KEY"), "The Cloud API Key")
		clusterID = flag.String("cluster-id", os.Getenv("CLOUD_CLUSTER_ID"), "The Cloud Cluster ID")
	)
	klog.InitFlags(nil)
	flag.Parse()

	if *apiURL == "" {
		*apiURL = fmt.Sprintf("https://thecloud-api.kube-system.svc.cluster.local:%d", defaultAPIPort)
	}
	if *apiKey == "" || *clusterID == "" {
		klog.Fatalf("API Key and Cluster ID are required")
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", *port))
	if err != nil {
		klog.Fatalf("failed to listen: %v", err)
	}

	client := sdk.NewClient(*apiURL, *apiKey)
	server := autoscaler.NewAutoscalerServer(client, *clusterID)

	s := grpc.NewServer()
	protos.RegisterCloudProviderServer(s, server)

	klog.Infof("Autoscaler gRPC server listening on %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		klog.Fatalf("failed to serve: %v", err)
	}
}
