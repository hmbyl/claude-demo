package server

import (
	"demo/internal/conf"
	"demo/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"

	pb "demo/api/helloworld/v1"
	pbAuth "demo/api/auth/v1"
)

// NewHTTPServer creates a new HTTP server.
func NewHTTPServer(c *conf.Server, greeter *service.GreeterService, auth *service.AuthService, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
		),
	}
	if c.HTTP.Addr != "" {
		opts = append(opts, http.Address(c.HTTP.Addr))
	}
	if c.HTTP.Timeout != 0 {
		opts = append(opts, http.Timeout(c.HTTP.Timeout))
	}
	srv := http.NewServer(opts...)
	pb.RegisterGreeterHTTPServer(srv, greeter)
	pbAuth.RegisterAuthHTTPServer(srv, auth)
	return srv
}
