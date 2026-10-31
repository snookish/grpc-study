package main

import (
	"context"
	"flag"
	"log"
	"time"

	personv1 "github.com/snookish/grpc/gen/go/person/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func main() {
	addr := flag.String("addr", "localhost:50051", "server address")
	flag.Parse()

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	client := personv1.NewPersonServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	saved, err := client.SavePerson(ctx, &personv1.SavePersonRequest{
		Name: "Jane Doe",
		Dob:  timestamppb.New(time.Date(2001, 9, 22, 0, 0, 0, 0, time.UTC)),
	})
	if err != nil {
		log.Fatalf("save person: %v", err)
	}
	log.Printf("saved: %s (dob: %s)", saved.GetPerson().GetName(), saved.GetPerson().GetDob().AsTime().Format("2006-01-02"))

	person, err := client.GetPerson(ctx, &personv1.GetPersonRequest{Name: "Jane Doe"})
	if err != nil {
		log.Fatalf("get person: %v", err)
	}
	log.Printf("person: %s (dob: %s)", person.GetPerson().GetName(), person.GetPerson().GetDob().AsTime().Format("2006-01-02"))

	people, err := client.ListPeople(ctx, &personv1.ListPeopleRequest{})
	if err != nil {
		log.Fatalf("list people: %v", err)
	}
	for _, p := range people.GetPeople() {
		log.Printf("listed: %s (dob: %s)", p.GetName(), p.GetDob().AsTime().Format("2006-01-02"))
	}
}
