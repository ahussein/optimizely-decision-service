package main

import (
	"context"
	"log"
	"net"

	pb "github.com/ahussein/optimizely-decision-service/internal/activate"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.ActivateServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) Activate(ctx context.Context, in *pb.ActivateRequest) (*pb.ActivateResponse, error) {
	return &pb.ActivateResponse{
		Result: "activated",
		Error:  "",
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterActivateServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
