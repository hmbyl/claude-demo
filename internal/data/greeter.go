package data

import (
	"context"
	"demo/internal/repo"

	"github.com/go-kratos/kratos/v2/log"
)

type greeterRepo struct {
	data *Data
	log  *log.Helper
}

// NewGreeterRepo creates a new GreeterRepo.
func NewGreeterRepo(data *Data, logger log.Logger) repo.GreeterRepo {
	return &greeterRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *greeterRepo) Save(ctx context.Context, g *repo.Greeter) (*repo.Greeter, error) {
	return g, nil
}

func (r *greeterRepo) FindByHello(ctx context.Context, hello string) (*repo.Greeter, error) {
	return &repo.Greeter{Hello: hello}, nil
}
