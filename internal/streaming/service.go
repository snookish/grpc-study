package streaming

import (
	"errors"
	"io"
	"log"
	"strings"
	"time"

	commonv1 "github.com/snookish/grpc/gen/go/common/v1"
	streamingv1 "github.com/snookish/grpc/gen/go/streaming/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	streamingv1.UnimplementedStreamingServiceServer
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) LogStream(stream grpc.ClientStreamingServer[streamingv1.LogStreamRequest, streamingv1.LogStreamResponse]) error {
	var count int32

	for {
		req, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return stream.SendAndClose(&streamingv1.LogStreamResponse{EntriesLogged: count})
			}
			return err
		}

		if err := validateLogRequest(req); err != nil {
			return err
		}

		log.Printf(
			"log received: level=%s timestamp=%s message=%q",
			req.GetLevel(),
			req.GetTimestamp().AsTime().Format(time.RFC3339),
			req.GetMessage(),
		)
		count++
	}
}

func (s *Service) StreamServerTime(
	req *streamingv1.StreamServerTimeRequest,
	stream grpc.ServerStreamingServer[streamingv1.StreamServerTimeResponse],
) error {
	interval := req.GetIntervalSeconds()
	if interval <= 0 {
		return status.Error(codes.InvalidArgument, "interval_seconds must be greater than 0")
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case <-ticker.C:
			if err := stream.Send(&streamingv1.StreamServerTimeResponse{
				CurrentTimestamp: timestamppb.Now(),
			}); err != nil {
				return err
			}
		}
	}
}

func validateLogRequest(req *streamingv1.LogStreamRequest) error {
	if req.GetTimestamp() == nil {
		return status.Error(codes.InvalidArgument, "timestamp is required")
	}
	if err := req.GetTimestamp().CheckValid(); err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid timestamp: %v", err)
	}
	if req.GetLevel() == commonv1.LogLevel_LOG_LEVEL_UNSPECIFIED {
		return status.Error(codes.InvalidArgument, "level is required")
	}
	if strings.TrimSpace(req.GetMessage()) == "" {
		return status.Error(codes.InvalidArgument, "message is required")
	}
	return nil
}
