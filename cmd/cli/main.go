package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"redis-clone/proto/pb"

	"github.com/google/shlex"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to dial grpc at port :50051: %v", err)
	}
	defer conn.Close()

	c := pb.NewKeyValueStoreClient(conn)
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("localhost:50051> ")
		input, _ := reader.ReadString('\n')

		tokens, err := shlex.Split(input)
		if err != nil {
			log.Printf("syntax error: %v", err)
			continue
		}
		if len(tokens) == 0 {
			continue
		}

		command := strings.ToUpper(tokens[0])
		args := tokens[1:]

		switch command {
		case "SET":
			if len(args) < 2 {
				log.Println("error USAGE: SET <key> <value>")
				continue
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

			_, err = c.Set(ctx, &pb.SetRequest{Key: args[0], Value: args[1]})
			cancel()
			if err != nil {
				log.Printf("SET failed: %v\n", err)
				continue
			}
			log.Printf("successfully added Key: %v, Value: %v", args[0], args[1])
		case "GET":
			if len(args) < 1 || len(args) > 1 {
				log.Println("error USAGE: GET <key>")
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

			resp, err := c.Get(ctx, &pb.GetRequest{Key: args[0]})
			cancel()
			if err != nil {
				log.Printf("GET failed error: %v", err)
				continue
			}
			log.Printf("%v : %v", args[0], resp)
		case "DELETE":
			if len(args) < 1 {
				log.Print("error USAGE: DELETE <key>")
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

			resp, err := c.Delete(ctx, &pb.DeleteRequest{Key: args[0]})
			cancel()
			if err != nil {
				log.Printf("DELETE error: %v", err)
				continue
			}
			log.Printf("%v \n", resp)
		case "BFADD":
			if len(args) < 1 {
				log.Println("error USAGE: BFADD <key>")
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

			resp, err := c.BFAdd(ctx, &pb.BFAddRequest{Key: args[0]})
			cancel()
			if err != nil {
				log.Printf("bloom filter add failed error: %v", err)
				continue
			}
			log.Printf("%v", resp)
		case "BFEXISTS":
			if len(args) < 1 {
				log.Println("invalid USAGE: BFEXISTS <key>")
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

			resp, err := c.BFExists(ctx, &pb.BFExistsRequest{Key: args[0]})
			cancel()
			if err != nil {
				log.Printf("bloom filter exists cmd error: %v", err)
				continue
			}
			log.Printf("%v", resp)

		default:
			log.Printf("ERR: Unknown command %q", command)
		}
	}
}
