package main

import (
	"context"
	"log"
	"strconv"

	pb "template/routeguide"

	"google.golang.org/grpc"
)

var client pb.ServiceClient

func main() {
	connection := connect(5000)
	if connection == nil {
		log.Fatalln("Connection failed")
		return
	}
	defer connection.Close()
	RequestMessage()
}

func connect(port int) *grpc.ClientConn {
	address := "localhost:" + strconv.Itoa(port)
	connection, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Println(address, "not active -", err)
		return nil
	}
	_client := pb.NewServiceClient(connection)
	client = _client
	return connection
}

func RequestMessage() {
	ctx := context.Background()
	response, err := client.MessageRPC(ctx, &pb.Empty{})
	if err != nil {
		log.Fatalln("Request failed", err)
		return
	}
	log.Println(response.Message)
}
