package main

import (
	"context"
	"log"
	"time"

	"redis-clone/proto/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to dial grpc server at port 50051: %v", err)
	}
	defer conn.Close()

	c := pb.NewKeyValueStoreClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = c.Set(ctx, &pb.SetRequest{Key: "test", Value: "hello grpc"})
	if err != nil {
		log.Fatalf("Set failed: %v\n", err)
	}
	log.Println("success")

	getResp, err := c.Get(ctx, &pb.GetRequest{Key: "test"})
	if err != nil {
		log.Fatalf("Get failed: %v\n", err)
	}
	log.Printf(" GET value: '%s' (exists: %v)\n", getResp.Value, getResp.Exists)
}
