package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"redis-clone/pkg/logger"

	"redis-clone/proto/pb"

	"github.com/chzyer/readline"
	"github.com/google/shlex"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("%s failed to dial grpc at port :50051: %v%s", logger.Color.Red(), err, logger.Color.Reset())
		return
	}
	defer conn.Close()

	c := pb.NewKeyValueStoreClient(conn)
	rl, err := readline.New("localhost:500051> ")
	if err != nil {
		log.Printf("%s failed to create new readline ERROR: %v%s", logger.Color.Red(), err, logger.Color.Reset())
		return
	}

	defer rl.Close()
	for {
		fmt.Print("localhost:50051> ")
		input, err := rl.Readline()
		if err != nil {
			break
		}

		tokens, err := shlex.Split(input)
		if err != nil {
			log.Printf("%s\u200Bsyntax error: %v%s", logger.Color.Red(), err, logger.Color.Reset())
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
				log.Printf("%s\u200Berror USAGE: SET <key> <value> [ttl]%s", logger.Color.Red(), logger.Color.Reset())
				continue
			}

			var ttl int64
			if len(args) == 3 {
				parsedTTL, err := strconv.ParseInt(args[2], 10, 64)
				if err != nil {
					log.Printf("%s\u200Binvalid ttl must of type int (milliseconds)%s", logger.Color.Red(), logger.Color.Reset())
					continue
				}
				ttl = parsedTTL
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

			_, err = c.Set(ctx, &pb.SetRequest{Key: args[0], Value: args[1], Ttl: ttl})
			cancel()
			if err != nil {
				log.Printf("%s\u200BSET failed: %v\n%s", logger.Color.Red(), err, logger.Color.Reset())
				continue
			}
			log.Printf("%s\u200Bsuccessfully added Key: %s%v%s, Value: %s%v%s", logger.Color.Green(), logger.Color.Blue(), args[0], logger.Color.Green(), logger.Color.Blue(), args[1], logger.Color.Reset())
		case "GET":
			if len(args) < 1 || len(args) > 1 {
				log.Printf("%s\u200Berror USAGE: GET <key>%s", logger.Color.Red(), logger.Color.Reset())
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

			resp, err := c.Get(ctx, &pb.GetRequest{Key: args[0]})
			cancel()
			if err != nil {
				log.Printf("%sGET failed error: %v%s", logger.Color.Red(), err, logger.Color.Reset())
				continue
			}
			log.Printf("%v : %v", args[0], resp)
		case "DELETE":
			if len(args) < 1 {
				log.Printf("%sError USAGE: DELETE <key>%s", logger.Color.Red(), logger.Color.Reset())
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

			resp, err := c.Delete(ctx, &pb.DeleteRequest{Key: args[0]})
			cancel()
			if err != nil {
				log.Printf("%sDELETE error: %v%s", logger.Color.Red(), err, logger.Color.Reset())
				continue
			}
			log.Printf("%v \n", resp)
		case "BFADD":
			if len(args) < 1 {
				log.Printf("%sError USAGE: BFADD <key>%s", logger.Color.Red(), logger.Color.Reset())
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

			resp, err := c.BFAdd(ctx, &pb.BFAddRequest{Key: args[0]})
			cancel()
			if err != nil {
				log.Printf("%sBloom filter add failed error: %v%s", logger.Color.Red(), err, logger.Color.Reset())
				continue
			}
			log.Printf("%v", resp)
		case "BFEXISTS":
			if len(args) < 1 {
				log.Printf("%sInvalid USAGE: BFEXISTS <key>%s", logger.Color.Red(), logger.Color.Reset())
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

			resp, err := c.BFExists(ctx, &pb.BFExistsRequest{Key: args[0]})
			cancel()
			if err != nil {
				log.Printf("%sBloom filter exists cmd error: %v%s", logger.Color.Red(), err, logger.Color.Reset())
				continue
			}
			log.Printf("%v", resp)
		case "BGREWRITEAOF":
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			resp, err := c.BGRewrite(ctx, &pb.BGRewriteRequest{})
			cancel()
			if err != nil {
				log.Printf("%sBGREWRITEAOF failed: %v%s", logger.Color.Red(), err, logger.Color.Reset())
				continue
			}
			log.Printf("%v", resp.Status)
		case "CLEAR":
			log.Printf("\033c")
		case "HSET":
			if len(args) < 3 {
				log.Printf("%sError USAGE: HSET <key> <field> <value>%s", logger.Color.Red(), logger.Color.Reset())
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			_, err = c.HSet(ctx, &pb.HSetRequest{Key: args[0], Field: args[1], Value: strings.Join(args[2:], " ")})
			cancel()
			if err != nil {
				log.Printf("%sHSET failed: %v%s", logger.Color.Red(), err, logger.Color.Reset())
				continue
			}
			log.Printf("%sSuccessfully added, hash key: %s%v%s%s field: %s%v%s%s", logger.Color.Green(), logger.Color.Blue(), args[0], logger.Color.Reset(), logger.Color.Green(), logger.Color.Blue(), args[1], logger.Color.Red(), logger.Color.Reset())
		case "HGET":
			if len(args) != 2 {
				log.Println("error USAGE: HGET <key> <field>")
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			resp, err := c.HGet(ctx, &pb.HGetRequest{Key: args[0], Field: args[1]})
			cancel()
			if err != nil {
				log.Printf("%sHGET failed error: %v%s", logger.Color.Red(), err, logger.Color.Reset())
				continue
			}
			log.Printf("%v[%v] : %v", args[0], args[1], resp.Value)
		case "INFO":
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			resp, err := c.Info(ctx, &pb.InfoRequest{})
			cancel()
			if err != nil {
				log.Printf("%sINFO failed: %v%s", logger.Color.Red(), err, logger.Color.Reset())
				continue
			}
			log.Printf("\n"+
				"┌──────────────────────┬──────────────────────┐\n"+
				"│ Metric               │ Value                │\n"+
				"├──────────────────────┼──────────────────────┤\n"+
				"│ Keys Loaded          │ %-20d │\n"+
				"│ Cache Hits           │ %-20d │\n"+
				"│ Cache Misses         │ %-20d │\n"+
				"│ Evictions            │ %-20d │\n"+
				"└──────────────────────┴──────────────────────┘",
				resp.Keys, resp.Hits, resp.Misses, resp.Evictions)
		case "PUBLISH":
			if len(args) < 2 {
				log.Println("error USAGE: PUBLISH <channel> <message>")
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			resp, err := c.Publish(ctx, &pb.PublishRequest{Channel: args[0], Message: strings.Join(args[1:], " ")})
			cancel()
			if err != nil {
				log.Printf("PUBLISH failed: %v\n", err)
				continue
			}
			log.Printf("message published to %d subscribers", resp.SubscribersReached)
		case "SUBSCRIBE":
			if len(args) != 1 {
				log.Println("error USAGE: SUBSCRIBE <channel>")
				continue
			}

			stream, err := c.Subscribe(context.Background(), &pb.SubscribeRequest{Channel: args[0]})
			if err != nil {
				log.Printf("SUBSCRIBE failed: %v\n", err)
				continue
			}

			log.Printf("subed to %s waiting for a response (ctrl+c to exit)", args[0])
			for {
				msg, err := stream.Recv()
				if err != nil {
					log.Printf("Disconnected from channel: %v", err)
					break
				}
				log.Printf("[%s]: %s", msg.Channel, msg.Message)
			}
		default:
			log.Printf("ERR: unknown command %q", command)
		}

	}
}
