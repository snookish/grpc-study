package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	commonv1 "github.com/snookish/grpc/gen/go/common/v1"
	streamingv1 "github.com/snookish/grpc/gen/go/streaming/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func main() {
	addr := flag.String("addr", "localhost:50051", "server address")
	mode := flag.String("mode", "logs", "streaming mode: logs or time")
	interval := flag.Int64("interval", 1, "server time interval in seconds")
	flag.Parse()

	if *mode != "logs" && *mode != "time" {
		log.Fatalf("unknown mode %q: expected logs or time", *mode)
	}
	if *mode == "time" && *interval <= 0 {
		log.Fatal("interval must be greater than 0")
	}

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	client := streamingv1.NewStreamingServiceClient(conn)

	switch *mode {
	case "logs":
		runLogStream(client)
	case "time":
		runServerTimeStream(client, *interval)
	default:
		log.Fatalf("unknown mode %q: expected logs or time", *mode)
	}
}

func runLogStream(client streamingv1.StreamingServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stream, err := client.LogStream(ctx)
	if err != nil {
		log.Fatalf("open log stream: %v", err)
	}

	for i := 1; i <= 5; i++ {
		if err := stream.Send(&streamingv1.LogStreamRequest{
			Timestamp: timestamppb.Now(),
			Level:     commonv1.LogLevel_LOG_LEVEL_INFO,
			Message:   fmt.Sprintf("sample log %d", i),
		}); err != nil {
			log.Fatalf("send log: %v", err)
		}
		time.Sleep(200 * time.Millisecond)
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("close log stream: %v", err)
	}
	log.Printf("logs sent: %d", resp.GetEntriesLogged())
}

func runServerTimeStream(client streamingv1.StreamingServiceClient, interval int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := client.StreamServerTime(ctx, &streamingv1.StreamServerTimeRequest{
		IntervalSeconds: interval,
	})
	if err != nil {
		log.Fatalf("open time stream: %v", err)
	}

	for {
		resp, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) || ctx.Err() != nil {
				return
			}
			log.Fatalf("receive time: %v", err)
		}
		log.Printf("server time: %s", resp.GetCurrentTimestamp().AsTime().Format(time.RFC3339))
	}
}
