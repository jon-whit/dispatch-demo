package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/authzed/consistent"
	"github.com/cespare/xxhash"
	"github.com/jon-whit/dispatch-demo/dispatch"
	cache "github.com/jon-whit/dispatch-demo/dispatch/cached"
	"github.com/jon-whit/dispatch-demo/dispatch/local"
	"github.com/jon-whit/dispatch-demo/dispatch/peer"
	dispatchv1 "github.com/jon-whit/dispatch-demo/proto/dispatch/v1"
	fgav1 "github.com/jon-whit/dispatch-demo/proto/fga/v1"
	"github.com/sercand/kuberesolver/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthv1pb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type fgaV1Service struct {
	fgav1.UnimplementedFGAServiceServer

	dispatcher dispatch.Dispatcher
}

func (f *fgaV1Service) Check(
	ctx context.Context,
	req *fgav1.CheckRequest,
) (*fgav1.CheckResponse, error) {
	fmt.Println("Check has been called")

	resp, err := f.dispatcher.DispatchCheck(ctx, &dispatchv1.DispatchCheckRequest{
		ObjectType: req.GetObjectType(),
		ObjectId:   req.GetObjectId(),
		Relation:   req.GetRelation(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to dispach Check request: %v", err)
	}

	fmt.Printf("DispatchCheck response '%v'\n", resp.GetAllowed())

	return &fgav1.CheckResponse{}, nil
}

type dispatchV1Service struct {
	dispatchv1.UnimplementedDispatchServiceServer

	localDispatcher dispatch.Dispatcher
}

func (d *dispatchV1Service) DispatchCheck(
	ctx context.Context,
	req *dispatchv1.DispatchCheckRequest,
) (*dispatchv1.DispatchCheckResponse, error) {

	fmt.Printf("calling localDispatcher.DispatchCheck '%s:%s#%s'\n", req.GetObjectType(), req.GetObjectId(), req.GetRelation())

	start := time.Now()
	defer func() {
		fmt.Println(time.Since(start))
	}()

	return d.localDispatcher.DispatchCheck(ctx, req)
}

func main() {

	// Register kuberesolver to grpc before calling grpc.Dial
	kuberesolver.RegisterInCluster()

	balancer.Register(consistent.NewBuilder(xxhash.Sum64))

	conn, err := grpc.Dial(
		"kubernetes:///fga:dispatcher-grpc",
		grpc.WithDefaultServiceConfig(consistent.DefaultServiceConfigJSON),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("grpc client connection failed: %v", err)
	}
	defer conn.Close()

	dispatchClient := dispatchv1.NewDispatchServiceClient(conn)

	peerDispatcher := &peer.PeerDispatcher{
		DispatchClient: dispatchClient,
	}

	localDispatcher := &local.LocalDispatcher{}
	//localDispatcher.Delegate = localDispatcher - if you want to loopback locally
	//localDispatcher.Delegate = peerDispatcher // if you want to delegate to a remote peer directly

	// todo: localDispatcher should be looped back with cachedLocalDispatcher because
	// overlapping subproblems on the same object id shouldn't have to incurr a network
	// hop

	cachedLocalDispatcher := &cache.CachedDispatcher{
		Cache: map[string]*dispatchv1.DispatchCheckResponse{
			"document:1#editor": {
				Allowed: true,
			},
		},
	}

	localDispatcher.Delegate = cachedLocalDispatcher

	cachedLocalDispatcher.Delegate = peerDispatcher

	dispatchService := &dispatchV1Service{
		localDispatcher: localDispatcher,
	}

	fgaServer := grpc.NewServer()
	healthServer := health.NewServer()

	fgav1.RegisterFGAServiceServer(fgaServer, &fgaV1Service{
		dispatcher: cachedLocalDispatcher,
	})
	healthv1pb.RegisterHealthServer(fgaServer, healthServer)
	reflection.Register(fgaServer)

	dispatchServer := grpc.NewServer()
	dispatchv1.RegisterDispatchServiceServer(dispatchServer, dispatchService)
	reflection.Register(dispatchServer)

	fgaListener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	dispatchListener, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	go func() {
		if err := dispatchServer.Serve(dispatchListener); err != nil {
			if !errors.Is(err, grpc.ErrServerStopped) {
				log.Fatalf("failed to start grpc server: %v", err)
			}
		}
	}()

	go func() {
		if err := fgaServer.Serve(fgaListener); err != nil {
			if !errors.Is(err, grpc.ErrServerStopped) {
				log.Fatalf("failed to start grpc server: %v", err)
			}
		}
	}()

	healthServer.SetServingStatus(fgav1.FGAService_ServiceDesc.ServiceName, healthv1pb.HealthCheckResponse_SERVING)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-done
}
