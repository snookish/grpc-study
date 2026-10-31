package person

import (
	"context"
	"testing"
	"time"

	personv1 "github.com/snookish/grpc/gen/go/person/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestSaveAndGetPerson(t *testing.T) {
	svc := NewService()
	ctx := context.Background()
	dob := timestamppb.New(time.Date(2001, 9, 22, 0, 0, 0, 0, time.UTC))

	saved, err := svc.SavePerson(ctx, &personv1.SavePersonRequest{
		Name: "  Jane Doe  ",
		Dob:  dob,
	})
	if err != nil {
		t.Fatalf("SavePerson returned error: %v", err)
	}
	if saved.GetPerson().GetName() != "Jane Doe" {
		t.Fatalf("SavePerson name = %q, want %q", saved.GetPerson().GetName(), "Jane Doe")
	}

	got, err := svc.GetPerson(ctx, &personv1.GetPersonRequest{Name: "Jane Doe"})
	if err != nil {
		t.Fatalf("GetPerson returned error: %v", err)
	}
	if got.GetPerson().GetName() != "Jane Doe" {
		t.Fatalf("GetPerson name = %q, want %q", got.GetPerson().GetName(), "Jane Doe")
	}
}

func TestSavePersonValidation(t *testing.T) {
	svc := NewService()
	ctx := context.Background()

	tests := []struct {
		name string
		req  *personv1.SavePersonRequest
		want codes.Code
	}{
		{
			name: "missing name",
			req:  &personv1.SavePersonRequest{Dob: timestamppb.Now()},
			want: codes.InvalidArgument,
		},
		{
			name: "missing dob",
			req:  &personv1.SavePersonRequest{Name: "Jane Doe"},
			want: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.SavePerson(ctx, tt.req)
			if got := status.Code(err); got != tt.want {
				t.Fatalf("SavePerson error code = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPersonNotFound(t *testing.T) {
	svc := NewService()

	_, err := svc.GetPerson(context.Background(), &personv1.GetPersonRequest{Name: "missing"})
	if got := status.Code(err); got != codes.NotFound {
		t.Fatalf("GetPerson error code = %v, want %v", got, codes.NotFound)
	}
}

func TestListPeopleSorted(t *testing.T) {
	svc := NewService()
	ctx := context.Background()
	dob := timestamppb.New(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))

	for _, name := range []string{"Charlie", "Alice", "Bob"} {
		if _, err := svc.SavePerson(ctx, &personv1.SavePersonRequest{Name: name, Dob: dob}); err != nil {
			t.Fatalf("SavePerson(%q) returned error: %v", name, err)
		}
	}

	resp, err := svc.ListPeople(ctx, &personv1.ListPeopleRequest{})
	if err != nil {
		t.Fatalf("ListPeople returned error: %v", err)
	}

	got := make([]string, 0, len(resp.GetPeople()))
	for _, p := range resp.GetPeople() {
		got = append(got, p.GetName())
	}

	want := []string{"Alice", "Bob", "Charlie"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ListPeople names = %v, want %v", got, want)
		}
	}
}
