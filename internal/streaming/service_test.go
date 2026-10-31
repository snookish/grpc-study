package streaming

import (
	"testing"

	commonv1 "github.com/snookish/grpc/gen/go/common/v1"
	streamingv1 "github.com/snookish/grpc/gen/go/streaming/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestValidateLogRequest(t *testing.T) {
	tests := []struct {
		name string
		req  *streamingv1.LogStreamRequest
		want codes.Code
	}{
		{
			name: "valid",
			req: &streamingv1.LogStreamRequest{
				Timestamp: timestamppb.Now(),
				Level:     commonv1.LogLevel_LOG_LEVEL_INFO,
				Message:   "hello",
			},
			want: codes.OK,
		},
		{
			name: "missing timestamp",
			req: &streamingv1.LogStreamRequest{
				Level:   commonv1.LogLevel_LOG_LEVEL_INFO,
				Message: "hello",
			},
			want: codes.InvalidArgument,
		},
		{
			name: "invalid timestamp",
			req: &streamingv1.LogStreamRequest{
				Timestamp: &timestamppb.Timestamp{Seconds: 1, Nanos: 1_000_000_000},
				Level:     commonv1.LogLevel_LOG_LEVEL_INFO,
				Message:   "hello",
			},
			want: codes.InvalidArgument,
		},
		{
			name: "missing level",
			req: &streamingv1.LogStreamRequest{
				Timestamp: timestamppb.Now(),
				Message:   "hello",
			},
			want: codes.InvalidArgument,
		},
		{
			name: "missing message",
			req: &streamingv1.LogStreamRequest{
				Timestamp: timestamppb.Now(),
				Level:     commonv1.LogLevel_LOG_LEVEL_INFO,
				Message:   "   ",
			},
			want: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLogRequest(tt.req)
			if got := status.Code(err); got != tt.want {
				t.Fatalf("validateLogRequest error code = %v, want %v", got, tt.want)
			}
		})
	}
}
