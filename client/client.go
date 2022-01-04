package main

import (
	"context"
	"log"
	"strconv"
	"time"

	pb "template/routeguide"

	"google.golang.org/grpc"
)

var client pb.IncrementServiceClient
var frontendPorts = [2]int{3000, 3001}

func main() {
	connection := findActiveFrontend()
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
	_client := pb.NewIncrementServiceClient(connection)
	client = _client
	return connection
}

func findActiveFrontend() *grpc.ClientConn {
	client = nil
	for i := 0; i < len(frontendPorts); i++ {
		connection := connect(frontendPorts[i])
		log.Printf("Checking port %d", frontendPorts[i])
		_, err := client.AliveCheck(context.Background(), &pb.Timestamp{})
		if err != nil {
			log.Printf("Frontend with port: %d not running", frontendPorts[i])
			log.Println(err)
		} else {
			log.Printf("Found connection with port: %d", frontendPorts[i])
			return connection
		}
	}
	log.Fatalln("No active frontends")
	return nil
}

func Increment() {
	response, err := client.Increment(context.Background(), &pb.Timestamp{})
	if err != nil {
		log.Println("Connection lost. Finding new frontend")
		connection := findActiveFrontend()
		if connection == nil {
			log.Fatalln("No active frontends")
			return
		}

		log.Println("Found new frontend")

		Increment()
	} else {
		log.Printf("Incremented.. New value: %d", response.Value)
		time.Sleep(2 * time.Second)
		Increment()
	}
}
