package main

import (
	"context"
	"errors"
	"log"
	"math/rand"
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
var serverID int32
var leaderPort int32
var lastRequestTimeFromLeader time.Time

type BullyService struct {
	pb.UnimplementedBullyServiceServer
}

func main() {
	log.Println("Starting Bully service by the team Fiji (Philip Kristian Møller Flyvholm, Tue Edmund Gyhrs and Thor Tudal Lauridsen")
	leaderPort = -1
	setupServerPort()
	time.Sleep(10 * time.Second)
	go runServer()
	time.Sleep(2 * time.Second)
	go runClient()
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
	serverID = int32(_serverID)
}

func runServer() {
	log.Println("--- SERVER APP ---")

	address_string := "localhost:" + strconv.Itoa(int(listenPort))
	lis, err := net.Listen("tcp", address_string)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()

	pb.RegisterBullyServiceServer(s, &BullyService{})
	log.Printf("Server listening at %v", lis.Addr())
	serveErr := s.Serve(lis)

	if serveErr != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func runClient() {
	rSleep()
	if !HasLeader() {
		StartElection()
	}
	go sendHeartbeat()
}

func sendHeartbeat() {
	if leaderPort == listenPort {
		for _, v := range serverPorts {
			if v == int(listenPort) {
				lastRequestTimeFromLeader = time.Now()
				continue
			}
			//send heartbeat
			address := "localhost:" + strconv.Itoa(v)
			_connection, err := grpc.Dial(address, grpc.WithInsecure())
			if err != nil {
				log.Fatalf("Failed to connect: %v", err)
			}
			defer _connection.Close()

			bullyServiceClient := pb.NewBullyServiceClient(_connection)
			_, hearbeatErr := bullyServiceClient.SendHeartbeat(context.Background(), &pb.Empty{})
			if hearbeatErr != nil {
				continue
			}
		}

		time.Sleep(4 * time.Second)
	} else {
		var timeDiff = time.Since(lastRequestTimeFromLeader)
		if timeDiff > (7*time.Second) && leaderPort != -1 {
			log.Println("Leader maybe dead. Starting election")
			leaderPort = -1
			StartElection()
		}
		time.Sleep(7 * time.Second)
	}
	sendHeartbeat()
}

func rSleep() {
	log.Printf("Beginning sleep")
	rand.Seed(time.Now().UnixNano())
	n := time.Duration(rand.Intn(5000)) * time.Millisecond //gets a random time between 0-5000 milliseonds (0-5 seconds)
	log.Printf("Sleeping %d seconds...\n", int(n))
	time.Sleep(n)
}

func StartElection() {
	if leaderPort == listenPort {
		return
	}
	var foundHigherServerToCoordinateElection bool
	for i := int(serverID); i < len(serverPorts); i++ {
		if i == int(serverID) {
			continue
		}
		address := "localhost:" + strconv.Itoa(serverPorts[i])
		_connection, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer _connection.Close()

		bullyServiceClient := pb.NewBullyServiceClient(_connection)
		response, err := bullyServiceClient.Election(context.Background(), &pb.ElectionRequest{Requester: listenPort})
		if err != nil {
			continue
		}
		if response != nil && response.Replier != -1 {
			foundHigherServerToCoordinateElection = true
		}
	}
	if !foundHigherServerToCoordinateElection {
		SetSelfAsLeader()
	}
}

func SetSelfAsLeader() {
	leaderPort = listenPort
	for i := 0; i < len(serverPorts); i++ {
		if i == int(serverID) {
			continue
		}
		address := "localhost:" + strconv.Itoa(serverPorts[i])
		_connection, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			continue
		}
		defer _connection.Close()
		bullyServiceClient := pb.NewBullyServiceClient(_connection)
		bullyServiceClient.Coordinator(context.Background(), &pb.CoordinatorRequest{Coordinator: listenPort})
		log.Printf("Sent coordinator to %s", address)
	}
	log.Printf("I am leader")
}

func HasLeader() bool {
	return leaderPort != -1
}

func (s *BullyService) Election(ctx context.Context, electionRequest *pb.ElectionRequest) (*pb.ElectionReply, error) {
	log.Printf("Recieved election request from %d. Timestamp: %d", electionRequest.Requester, timestamp.Increment())
	StartElection()
	return &pb.ElectionReply{Replier: listenPort}, nil
}

func (s *BullyService) Coordinator(ctx context.Context, coordinatorRequest *pb.CoordinatorRequest) (*pb.Empty, error) {
	leaderPort = coordinatorRequest.Coordinator
	lastRequestTimeFromLeader = time.Now()
	log.Println("Recieved coordinator request. Timestamp:", timestamp.Increment(), "New leader:", leaderPort)
	return &pb.Empty{}, nil
}

func (s *BullyService) AskForLeader(ctx context.Context, askRequest *pb.AskRequest) (*pb.LeaderPort, error) {
	log.Println("Recieved message request. Timestamp:", timestamp.Increment())
	if !HasLeader() {
		return nil, errors.New("no leader selected yet")
	}
	return &pb.LeaderPort{LeaderPort: leaderPort}, nil
}

func (s *BullyService) SendHeartbeat(ctx context.Context, empty *pb.Empty) (*pb.Empty, error) {
	log.Printf("Recieved heartbeat. Timestamp: %d", timestamp.Increment())
	lastRequestTimeFromLeader = time.Now()
	return &pb.Empty{}, nil
}

func (s *BullyService) ShareData(ctx context.Context, data *pb.Data) (*pb.Empty, error) {
	log.Println("Recieved message request. Timestamp:", timestamp.Increment())
	return &pb.Empty{}, nil
}
