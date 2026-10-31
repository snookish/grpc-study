package person

import (
	"context"
	"sort"
	"strings"
	"sync"

	personv1 "github.com/snookish/grpc/gen/go/person/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	personv1.UnimplementedPersonServiceServer

	mu     sync.RWMutex
	people map[string]*personv1.Person
}

func NewService() *Service {
	return &Service{people: make(map[string]*personv1.Person)}
}

func (s *Service) SavePerson(_ context.Context, req *personv1.SavePersonRequest) (*personv1.SavePersonResponse, error) {
	name := strings.TrimSpace(req.GetName())
	if name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if err := validateDOB(req.GetDob()); err != nil {
		return nil, err
	}

	p := &personv1.Person{Name: name, Dob: req.GetDob()}

	s.mu.Lock()
	s.people[name] = p
	s.mu.Unlock()

	return &personv1.SavePersonResponse{Person: p}, nil
}

func (s *Service) GetPerson(_ context.Context, req *personv1.GetPersonRequest) (*personv1.GetPersonResponse, error) {
	name := strings.TrimSpace(req.GetName())

	s.mu.RLock()
	p, ok := s.people[name]
	s.mu.RUnlock()

	if !ok {
		return nil, status.Errorf(codes.NotFound, "person %q not found", name)
	}
	return &personv1.GetPersonResponse{Person: p}, nil
}

func (s *Service) ListPeople(context.Context, *personv1.ListPeopleRequest) (*personv1.ListPeopleResponse, error) {
	s.mu.RLock()
	people := make([]*personv1.Person, 0, len(s.people))
	for _, p := range s.people {
		people = append(people, p)
	}
	s.mu.RUnlock()

	sort.Slice(people, func(i, j int) bool {
		return people[i].GetName() < people[j].GetName()
	})

	return &personv1.ListPeopleResponse{People: people}, nil
}

func validateDOB(dob *timestamppb.Timestamp) error {
	if dob == nil {
		return status.Error(codes.InvalidArgument, "dob is required")
	}
	if err := dob.CheckValid(); err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid dob: %v", err)
	}
	return nil
}
