package main

import (
	"bytes"
	"context"
	"encoding/json"
	"grpc-golang/pb"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"google.golang.org/grpc"
)

func main() {
	connection, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Falha ao conectar: %v", err)
	}

	defer connection.Close()

	// notes := []*pb.Note{
	// 	{Message: "First message"},
	// 	{Message: "Second message"},
	// 	{Message: "Third message"},
	// 	{Message: "Fourth message"},
	// 	{Message: "Fifth message"},
	// 	{Message: "Sixth message"},
	// }

	notes := []*pb.Note{
		{Message: "Test message"},
	}

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	client := pb.NewHelloServiceClient(connection)
	runGRPCChat(client, notes)

	runRESTChat(notes)
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

func runGRPCChat(client pb.HelloServiceClient, notes []*pb.Note) {
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream, err := client.Chat(ctx)
	if err != nil {
		log.Fatalf("%v.RouteChat(_) = _, %v", client, err)
	}
	waitc := make(chan struct{})
	msgReceived := 0
	start := time.Now()

	go func() {
		for {
			_, err := stream.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive a note : %v", err)
			}
			// log.Printf("Got message %s", in.Message)
			msgReceived++

			if msgReceived >= 1000 {
				stream.CloseSend()
			}
		}
	}()
	// for _, note := range notes {
	for i := 0; i < 1000; i++ {
		if err := stream.Send(notes[0]); err != nil {
			log.Fatalf("Failed to send a note: %v", err)
		}
	}
	<-waitc
	log.Println(time.Since(start))
}

func runRESTChat(notes []*pb.Note) {
	waitg := sync.WaitGroup{}
	start := time.Now()

	for i := 0; i < 1000; i++ {
		waitg.Add(1)

		go func() {
			postTest(notes)
			waitg.Done()
		}()
	}
	waitg.Wait()
	log.Println(time.Since(start))
}

func postTest(notes []*pb.Note) {
	url := "http://localhost:3003"
	jsonValue, _ := json.Marshal(notes[0])

	_, err := http.Post(url, "json", bytes.NewBuffer(jsonValue))
	if err != nil {
		log.Println("falha", err)
	}

	// log.Println("response", response)
}
