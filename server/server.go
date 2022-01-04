package main

import (
	"context"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	pb "template/routeguide"
	sh "template/shared"

	"google.golang.org/grpc"
)

var timestamp sh.SafeTimestamp
var listenPort int32
var serverPorts = [3]int{5000, 5001, 5002}

var num int32

type IncrementService struct {
	pb.UnimplementedIncrementServiceServer
}

func main() {
	log.Println("Starting increment service by the team Fiji (Philip Kristian Møller Flyvholm, Tue Edmund Gyhrs and Thor Tudal Lauridsen)")
	num = -1
	setupServerPort()
	go runServer()
	for {
		time.Sleep(2 * time.Second)
	}
}

func setupServerPort() {
	if len(os.Args) == 1 {
		log.Println("Please choose a server between 0 and", len(serverPorts)-1)
		return
	}
	_serverID, err1 := strconv.Atoi(os.Args[1])
	if err1 != nil {
		log.Fatalf("Bad serverId")
	}
	listenPort = int32(serverPorts[_serverID])
}

func runServer() {
	log.Println("--- SERVER APP ---")

	address_string := "localhost:" + strconv.Itoa(int(listenPort))
	lis, err := net.Listen("tcp", address_string)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()

	pb.RegisterIncrementServiceServer(s, &IncrementService{})
	log.Printf("Server listening at %v", lis.Addr())
	serveErr := s.Serve(lis)

	if serveErr != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func (s *IncrementService) AliveCheck(ctx context.Context, askRequest *pb.Timestamp) (*pb.Timestamp, error) {
	log.Println("Recieved alive request. Timestamp:", timestamp.MaxInc(askRequest.Time))

	return &pb.Timestamp{Time: timestamp.Value()}, nil
}

func (s *IncrementService) Increment(ctx context.Context, data *pb.Timestamp) (*pb.Value, error) {
	log.Println("Recieved increment request. Timestamp:", timestamp.MaxInc(data.Time))
	num++
	return &pb.Value{Value: num, Time: timestamp.Value()}, nil
}
