package main

import (
	"context"
	"log"
	"net"
	"strconv"

	pb "template/routeguide"
	sh "template/shared"

	"google.golang.org/grpc"
)

var timestamp sh.SafeTimestamp
var listenPort = 5000

type Service struct {
	pb.UnimplementedServiceServer
}

func main() {
	runServer()
}

func runServer() {
	log.Println("--- SERVER APP ---")

	address_string := "localhost:" + strconv.Itoa(int(listenPort))
	lis, err := net.Listen("tcp", address_string)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()

	pb.RegisterServiceServer(s, &Service{})
	log.Printf("Server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func (s *Service) MessageRPC(ctx context.Context, empty *pb.Empty) (*pb.Message, error) {
	log.Println("Recieved message request. Timestamp:", timestamp.Increment())
	return &pb.Message{Message: "Hello world!"}, nil
}
