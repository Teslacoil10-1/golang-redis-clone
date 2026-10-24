// Package grpcapi
package grpcapi

import (
	"context"
	"fmt"
	"time"

	"redis-clone/internal/aof"
	"redis-clone/internal/bloomfilter"
	"redis-clone/internal/store"
	"redis-clone/proto/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedKeyValueStoreServer
	DB *store.Store
	BF *bloomfilter.Filter
}

func (s *Server) Set(ctx context.Context, req *pb.SetRequest) (*pb.SetResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var cmd string
	if req.Ttl > 0 {
		expiresAt := time.Now().UnixNano() + (req.Ttl * int64(time.Millisecond))
		s.DB.SetWithTTL(req.Key, req.Value, expiresAt)
		cmd = fmt.Sprintf("SET %s %s %d", req.Key, req.Value, expiresAt)
	} else {
		s.DB.Set(req.Key, req.Value)
		cmd = fmt.Sprintf("SET %s %s", req.Key, req.Value)
	}

	aof.RecordCommand(cmd)

	return &pb.SetResponse{Success: true}, nil
}

func (s *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	val, exists := s.DB.Get(req.Key)
	if !exists {
		return &pb.GetResponse{Value: "", Exists: false}, status.Error(codes.NotFound, "key not found")
	}
	aof.RecordCommand("GET " + req.Key)
	return &pb.GetResponse{Value: val, Exists: true}, nil
}

func (s *Server) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	deleted := s.DB.Delete(req.Key)

	aof.RecordCommand("DELETE " + req.Key)

	return &pb.DeleteResponse{Success: deleted}, nil
}

func (s *Server) BFAdd(ctx context.Context, req *pb.BFAddRequest) (*pb.BFAddResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.BF.Add(req.Key)

	aof.RecordCommand("BFADD " + req.Key)

	return &pb.BFAddResponse{Success: true}, nil
}

func (s *Server) BFExists(ctx context.Context, req *pb.BFExistsRequest) (*pb.BFExistsResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	exists := s.BF.Exists(req.Key)

	aof.RecordCommand("BFEXISTS " + req.Key)

	return &pb.BFExistsResponse{Exists: exists}, nil
}

func (s *Server) BGRewrite(ctx context.Context, req *pb.BGRewriteRequest) (*pb.BGRewriteResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	err := aof.TriggerRewrite(s.DB)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to trigger rewrite: %v", err)
	}

	return &pb.BGRewriteResponse{Status: "background aof rewrite successful"}, nil
}
