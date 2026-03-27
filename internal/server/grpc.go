package server

import (
	"demo/internal/conf"
	"demo/internal/service"

	pb "demo/api/helloworld/v1"
	pbAuth "demo/api/auth/v1"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	kratosgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer creates a new gRPC server.
func NewGRPCServer(c *conf.Server, greeter *service.GreeterService, auth *service.AuthService, logger log.Logger) *kratosgrpc.Server {
	var opts = []kratosgrpc.ServerOption{
		kratosgrpc.Middleware(
			recovery.Recovery(),
		),
	}
	if c.GRPC.Addr != "" {
		opts = append(opts, kratosgrpc.Address(c.GRPC.Addr))
	}
	if c.GRPC.Timeout != 0 {
		opts = append(opts, kratosgrpc.Timeout(c.GRPC.Timeout))
	}
	srv := kratosgrpc.NewServer(opts...)
	pb.RegisterGreeterServer(srv, greeter)
	pbAuth.RegisterAuthServer(srv, auth)
	return srv
}
