package internal

import (
	"context"
	"errors"
	"fmt"

	"github.com/gerladeno/homie-core/internal/models"
	"github.com/gerladeno/homie-core/internal/storage"
	"github.com/gerladeno/homie-core/pkg/common"
	"github.com/sirupsen/logrus"
)

type Storage interface {
	SaveConfig(ctx context.Context, config *models.Config) error
	GetConfig(ctx context.Context, uuid string) (*models.Config, error)
	GetRegions(ctx context.Context) ([]*models.Region, error)
	UpsertRelation(ctx context.Context, relation *models.Relation) error
	ListRelated(ctx context.Context, uuid string, relation storage.Relation, limit, offset int64) ([]*models.Profile, error)
	ListMatches(ctx context.Context, uuid string, count int64) ([]*models.Profile, error)
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
	switch {
	case err == nil:
	case errors.Is(err, common.ErrConfigNotFound):
		return nil, common.ErrConfigNotFound
	default:
		err = fmt.Errorf("err getting config: %w", err)
		a.log.Debug(err)
		return nil, err
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

func (a *App) Like(ctx context.Context, uuid, targetUUID string, super bool) error {
	relationType := storage.Liked
	if super {
		relationType = storage.SuperLiked
	}
	relation := models.Relation{
		UUID:     uuid,
		Target:   targetUUID,
		Relation: int8(relationType),
	}
	if err := a.store.UpsertRelation(ctx, &relation); err != nil {
		return fmt.Errorf("err adding relation")
	}
	return nil
}

func (a *App) Dislike(ctx context.Context, uuid, targetUUID string) error {
	relationType := storage.Disliked
	relation := models.Relation{
		UUID:     uuid,
		Target:   targetUUID,
		Relation: int8(relationType),
	}
	if err := a.store.UpsertRelation(ctx, &relation); err != nil {
		return fmt.Errorf("err adding relation")
	}
	return nil
}

func (a *App) ListLikedProfiles(ctx context.Context, uuid string, limit, offset int64) ([]*models.Profile, error) {
	liked, err := a.store.ListRelated(ctx, uuid, storage.Liked, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("err getting list of liked: %w", err)
	}
	return liked, nil
}

func (a *App) ListDislikedProfiles(ctx context.Context, uuid string, limit, offset int64) ([]*models.Profile, error) {
	disliked, err := a.store.ListRelated(ctx, uuid, storage.Disliked, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("err getting list of disliked: %w", err)
	}
	return disliked, nil
}

func (a *App) GetMatches(ctx context.Context, uuid string, count int64) ([]*models.Profile, error) {
	matches, err := a.store.ListMatches(ctx, uuid, count)
	if err != nil {
		return nil, fmt.Errorf("err getting list of matches")
	}
	return matches, nil
}
