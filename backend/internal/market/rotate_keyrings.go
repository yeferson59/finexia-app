package market

import (
	"context"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

type keyringService interface {
	Rotate() error
}

type RotateKeyrings struct {
	keyringService keyringService
	log            logger.Logger
}

func NewRotateKeyrings(keyringService keyringService, log logger.Logger) *RotateKeyrings {
	return new(RotateKeyrings{
		keyringService: keyringService,
		log:            log,
	})
}

func (*RotateKeyrings) Name() string {
	return "rotate_keyrings"
}

func (rk *RotateKeyrings) Run(ctx context.Context) error {
	if err := rk.keyringService.Rotate(); err != nil {
		rk.log.Error(ctx, "failed to rotate keyrings", logger.Err(err))

		return err
	}

	rk.log.Info(ctx, "keyrings rotated successfully")

	return nil
}
