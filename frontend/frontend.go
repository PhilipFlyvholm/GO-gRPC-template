package main

import (
	"context"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	pb "template/routeguide"
	sh "template/shared"

	"google.golang.org/grpc"
)

var lamport sh.SafeTimestamp
var frontendPorts = [2]int{3000, 3001}
var serverPorts = [3]int{5000, 5001, 5002}
var replicas []pb.IncrementServiceClient
var listenPort int
var m sync.Mutex

type IncrementService struct {
	pb.UnimplementedIncrementServiceServer
}

func main() {
	go runFrontend()

	connectToServers()
	//BLOCK THREAD
	for {
		time.Sleep(2 * time.Second)
	}
}

func runFrontend() {
	log.Println("--- Frontend APP ---")
	setupFrontendPort()
	address_string := "localhost:" + strconv.Itoa(listenPort)
	lis, err := net.Listen("tcp", address_string)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()

	pb.RegisterIncrementServiceServer(s, &IncrementService{})
	log.Printf("Frontend listening at %v", lis.Addr())
	serveErr := s.Serve(lis)

	if serveErr != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func connectToServers() {
	for _, v := range serverPorts {
		go connectToServer(v)
	}
}

func connectToServer(port int) {
	address := "localhost:" + strconv.Itoa(port)
	connection, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Println(address, "not active -", err)
		return
	}
	defer connection.Close()
	_client := pb.NewIncrementServiceClient(connection)
	m.Lock()
	replicas = append(replicas, _client)
	m.Unlock()
	log.Printf("Connected to %d", port)
	//BLOCK THREAD
	for {
		time.Sleep(2 * time.Second)
	}
}

func setupFrontendPort() {
	if len(os.Args) == 1 {
		log.Println("Please choose a port between 0 and", len(frontendPorts)-1)
		return
	}
	_portID, err1 := strconv.Atoi(os.Args[1])
	if err1 != nil {
		log.Fatalf("Bad portId")
	}
	listenPort = frontendPorts[_portID]
}

func (s *IncrementService) AliveCheck(ctx context.Context, time *pb.Timestamp) (*pb.Timestamp, error) {
	log.Println("Recieved alive request. Timestamp:", lamport.MaxInc(time.Time))

	return &pb.Timestamp{Time: lamport.Value()}, nil
}

func (s *IncrementService) Increment(ctx context.Context, time *pb.Timestamp) (*pb.Value, error) {
	log.Println("Recieved increment request. Timestamp:", lamport.MaxInc(time.Time))

	responseMap := make(map[int32]int)
	var max int = 0
	var correctAmount int32
	for _, replica := range replicas {
		response, _ := replica.Increment(ctx, &pb.Timestamp{Time: lamport.Value()})

		if response != nil {

			log.Println(response.Value)

			responseMap[response.Value]++
			var currentCount = responseMap[response.Value]
			if max < currentCount {
				max = currentCount
				correctAmount = response.Value
			}
		}
	}

	if max > 1 {
		log.Printf("Found consensus on value: %d", correctAmount)
		return &pb.Value{Value: correctAmount}, nil
	}

	log.Println("Could not find consensus")
	return &pb.Value{Value: -1}, nil
}
