package main

import (
	"context"
	"log"
	"time"

	pb "github.com/ahussein/optimizely-decision-service/internal/activate"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	address     = "localhost:50051"
	defaultName = "world"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewActivateClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	experimentKey := "us-widget-bff"
	m := map[string]interface{}{
		"customer_uuid": "b5aedcf2-c5d8-4bd1-a4df-4d76702cea74",
		"country":       "US",
		"platform":      "mobile",
		"public_id":     "jhds",
	}

	attributes, err := structpb.NewStruct(m)
	if err != nil {
		panic(err)
	}

	user := &pb.User{
		Id:         "b5aedcf2-c5d8-4bd1-a4df-4d76702cea74",
		Attributes: attributes,
	}
	defer cancel()
	r, err := c.Activate(ctx, &pb.ActivateRequest{
		ExperimentKey: experimentKey,
		User:          user,
	})
	if err != nil {
		log.Fatalf("Could not get the activate variation", err)
	}
	log.Printf("Variation: %s", r.Variation)

}
