package grpcapi

import (
	"context"

	"redis-clone/internal/store"
	"redis-clone/proto/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedKeyValueStoreServer
	DB *store.Store
}

func (s *Server) Set(ctx context.Context, req *pb.SetRequest) (*pb.SetResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.DB.Set(req.Key, req.Value)
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

	return &pb.GetResponse{Value: val, Exists: true}, nil
}

func (s *Server) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	deleted, _ := s.DB.Delete(req.Key)
	return &pb.DeleteResponse{Success: deleted}, nil
}
