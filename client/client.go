package main

import (
	"context"
	"log"
	"strconv"
	"time"

	pb "template/routeguide"

	"google.golang.org/grpc"
)

var client pb.BullyServiceClient
var serverPorts = [3]int{5000, 5001, 5002}

func main() {
	connection := findActiveServer()
	defer connection.Close()
	Increment()
}

func connect(port int) *grpc.ClientConn {
	address := "localhost:" + strconv.Itoa(port)
	connection, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Println(address, "not active -", err)
		return nil
	}
	_client := pb.NewBullyServiceClient(connection)
	client = _client
	return connection
}

func findActiveServer() *grpc.ClientConn {
	client = nil
	for i := 0; i < len(serverPorts); i++ {
		connection := connect(serverPorts[i])
		log.Printf("Checking port %d", serverPorts[i])
		_, err := client.AliveCheck(context.Background(), &pb.Empty{})
		if err != nil {
			log.Printf("Server with port: %d not running", serverPorts[i])
			log.Println(err)
		} else {
			log.Printf("Found connection with port: %d", serverPorts[i])
			return connection
		}
	}
	log.Fatalln("No active servers")
	return nil
}

func Increment() {
	response, err := client.Increment(context.Background(), &pb.Empty{})
	if err != nil {
		log.Println("Connection lost. Finding new server")
		connection := findActiveServer()
		if connection == nil {
			log.Fatalln("No active servers")
			return
		}

		log.Println("Found new server")

		Increment()
	} else {
		log.Printf("Incremented.. New value: %d", response.Value)
		time.Sleep(2 * time.Second)
		Increment()
	}
}
