package service

import (
	"context"
	"demo/internal/biz"
	"demo/internal/repo"

	pb "demo/api/helloworld/v1"
)

// GreeterService implements the helloworld gRPC/HTTP service.
type GreeterService struct {
	pb.UnimplementedGreeterServer
	uc *biz.GreeterUsecase
}

// NewGreeterService creates a new GreeterService.
func NewGreeterService(uc *biz.GreeterUsecase) *GreeterService {
	return &GreeterService{uc: uc}
}

// SayHello handles the SayHello RPC.
func (s *GreeterService) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	g, err := s.uc.CreateGreeter(ctx, &repo.Greeter{Hello: in.Name})
	if err != nil {
		return nil, err
	}
	return &pb.HelloReply{Message: "Hello " + g.Hello}, nil
}
