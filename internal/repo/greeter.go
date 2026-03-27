package repo

import (
	"context"
)

// Greeter is a Greeter model.
type Greeter struct {
	Hello string
}

// GreeterRepo is a Greeter repo interface.
type GreeterRepo interface {
	Save(ctx context.Context, g *Greeter) (*Greeter, error)
	FindByHello(ctx context.Context, hello string) (*Greeter, error)
}
