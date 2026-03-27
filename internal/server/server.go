package server

import (
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewHTTPServer, NewGRPCServer)

// NewServers groups HTTP and gRPC servers (used by app initializer).
func NewServers(hs *http.Server, gs *grpc.Server) []interface{} {
	return []interface{}{hs, gs}
}
