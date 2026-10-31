package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	personv1 "github.com/snookish/grpc/gen/go/person/v1"
	streamingv1 "github.com/snookish/grpc/gen/go/streaming/v1"
	"github.com/snookish/grpc/internal/person"
	"github.com/snookish/grpc/internal/streaming"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	addr := flag.String("addr", ":50051", "address to listen on")
	flag.Parse()

	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	defer lis.Close()

	srv := grpc.NewServer()
	personv1.RegisterPersonServiceServer(srv, person.NewService())
	streamingv1.RegisterStreamingServiceServer(srv, streaming.NewService())

	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthv1.HealthCheckResponse_SERVING)
	healthv1.RegisterHealthServer(srv, healthServer)
	reflection.Register(srv)

	go func() {
		log.Printf("gRPC server listening on %s", *addr)
		if err := srv.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Fatalf("serve: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	log.Println("shutting down gRPC server")
	healthServer.SetServingStatus("", healthv1.HealthCheckResponse_NOT_SERVING)

	stopped := make(chan struct{})
	go func() {
		srv.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-time.After(5 * time.Second):
		log.Println("graceful shutdown timed out; forcing stop")
		srv.Stop()
	}
}
