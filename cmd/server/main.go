package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"redis-clone/internal/bloomfilter"
	"redis-clone/internal/grpcapi"
	"redis-clone/internal/store"
	"redis-clone/proto/pb"

	"google.golang.org/grpc"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	rdb := store.CreateStore()
	bf := bloomfilter.New(100000, 0.01)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen on port 50051: %v\n", err)
	}

	s := grpc.NewServer()
	grpcServer := &grpcapi.Server{DB: rdb, BF: bf}
	pb.RegisterKeyValueStoreServer(s, grpcServer)

	go func() {
		log.Printf("Server is listening on %v\n", lis.Addr())

		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v\n", err)
		}
	}()

	<-ctx.Done()
	log.Println("\n\n gracefully exiting")
	s.GracefulStop()
}
