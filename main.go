package main

import (
	"context"
	"grpc-golang/pb"
	"io"
	"log"

	"google.golang.org/grpc"
)

func main() {
	connection, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Falha ao conectar: %v", err)
	}

	defer connection.Close()

	client := pb.NewHelloServiceClient(connection)

	// Hello(client)
	runRouteChat(client)
}

func Hello(client pb.HelloServiceClient) {
	request := &pb.HelloRequest{
		Name: "Dudu",
	}

	res, err := client.Hello(context.Background(), request)
	if err != nil {
		log.Fatalf("Erro durante a execução: %v", err)
	}

	log.Println(res.Msg)
}

// runRouteChat receives a sequence of route notes, while sending notes for various locations.
func runRouteChat(client pb.HelloServiceClient) {
	notes := []*pb.Note{
		{Message: "First message"},
		{Message: "Second message"},
		{Message: "Third message"},
		{Message: "Fourth message"},
		{Message: "Fifth message"},
		{Message: "Sixth message"},
	}
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream, err := client.Chat(ctx)
	if err != nil {
		log.Fatalf("%v.RouteChat(_) = _, %v", client, err)
	}
	waitc := make(chan struct{})
	msgReceived := 0

	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive a note : %v", err)
			}
			log.Printf("Got message %s", in.Message)
			msgReceived++

			if msgReceived == len(notes) {
				stream.CloseSend()
			}
		}
	}()
	for _, note := range notes {
		if err := stream.Send(note); err != nil {
			log.Fatalf("Failed to send a note: %v", err)
		}
	}
	<-waitc
}
