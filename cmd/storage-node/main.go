// Package main provides the storage node entrypoint.
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"strings"
	"time"

	"github.com/poyrazk/thecloud/internal/storage/node"
	pb "github.com/poyrazk/thecloud/internal/storage/protocol"
	"google.golang.org/grpc"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	port := flag.String("port", "9101", "Port to listen on")
	dataDir := flag.String("data-dir", "./data/storage-node", "Directory to store data")
	peers := flag.String("peers", "", "Comma-separated list of peer addresses (e.g. localhost:9102)")
	nodeID := flag.String("id", "", "Unique Node ID (defaults to port)")
	flag.Parse()

	if *nodeID == "" {
		*nodeID = "node-" + *port
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("starting storage node", "id", *nodeID, "port", *port, "dataDir", *dataDir)

	// 1. Init Store
	store, err := node.NewLocalStore(*dataDir)
	if err != nil {
		logger.Error("failed to init store", "error", err)
		return err
	}

	// 2. Init Gossiper
	gossiper := node.NewGossipProtocol(*nodeID, "localhost:"+*port, logger)
	if *peers != "" {
		for _, peerAddr := range strings.Split(*peers, ",") {
			gossiper.AddPeer(peerAddr, peerAddr)
		}
	}
	gossiper.Start(1 * time.Second)
	defer gossiper.Stop()

	// 3. Init RPC Server
	rpcServer := node.NewRPCServer(store, gossiper)
	grpcServer := grpc.NewServer()
	pb.RegisterStorageNodeServer(grpcServer, rpcServer)

	// 3. Listen
	lis, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		logger.Error("failed to listen", "error", err)
		return err
	}

	// 4. Handle Shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		logger.Info("shutting down")
		grpcServer.GracefulStop()
	}()

	logger.Info("storage node ready")
	if err := grpcServer.Serve(lis); err != nil {
		logger.Error("server failed", "error", err)
		return err
	}
	return nil
}
