package biz

import (
	"context"
	"demo/internal/repo"

	"github.com/go-kratos/kratos/v2/log"
)

// GreeterUsecase is a Greeter usecase.
type GreeterUsecase struct {
	repo repo.GreeterRepo
	log  *log.Helper
}

// NewGreeterUsecase creates a GreeterUsecase.
func NewGreeterUsecase(repo repo.GreeterRepo, logger log.Logger) *GreeterUsecase {
	return &GreeterUsecase{repo: repo, log: log.NewHelper(logger)}
}

// CreateGreeter creates a Greeter and persists it.
func (uc *GreeterUsecase) CreateGreeter(ctx context.Context, g *repo.Greeter) (*repo.Greeter, error) {
	uc.log.WithContext(ctx).Infof("CreateGreeter: %v", g.Hello)
	return uc.repo.Save(ctx, g)
}
