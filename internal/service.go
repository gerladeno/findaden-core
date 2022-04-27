package internal

import (
	"context"
	"fmt"
	"github.com/gerladeno/homie-core/internal/models"
	"github.com/sirupsen/logrus"
)

type Storage interface {
	SaveConfig(ctx context.Context, config *models.Config) error
	GetConfig(ctx context.Context, uuid string) (*models.Config, error)
	GetRegions(ctx context.Context) ([]*models.Region, error)
}

type App struct {
	log   *logrus.Entry
	store Storage
}

func NewApp(log *logrus.Logger, store Storage) *App {
	return &App{
		log:   log.WithField("module", "app"),
		store: store,
	}
}

func (a *App) SaveConfig(ctx context.Context, config *models.Config) error {
	if err := a.store.SaveConfig(ctx, config); err != nil {
		return fmt.Errorf("err saving config: %w", err)
	}
	return nil
}
func (a *App) GetConfig(ctx context.Context, uuid string) (*models.Config, error) {
	result, err := a.store.GetConfig(ctx, uuid)
	if err != nil {
		return nil, fmt.Errorf("err getting config: %w", err)
	}
	return result, nil
}

func (a *App) GetRegions(ctx context.Context) ([]*models.Region, error) {
	result, err := a.store.GetRegions(ctx)
	if err != nil {
		return nil, fmt.Errorf("err getting regions: %w", err)
	}
	return result, nil
}
