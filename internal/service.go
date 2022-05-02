package internal

import (
	"context"
	"errors"
	"fmt"

	"github.com/gerladeno/homie-core/internal/models"
	"github.com/gerladeno/homie-core/internal/storage"
	"github.com/sirupsen/logrus"
)

var ErrConfigNotFound = errors.New("config not found")

type Storage interface {
	SaveConfig(ctx context.Context, config *models.Config) error
	GetConfig(ctx context.Context, uuid string) (*models.Config, error)
	GetRegions(ctx context.Context) ([]*models.Region, error)
	UpsertRelation(ctx context.Context, relation *models.Relation) error
	ListRelated(ctx context.Context, uuid string, relation storage.Relation, limit, offset int64) ([]*models.Config, error)
	ListUnrelated(ctx context.Context, uuid string, limit, offset int64) ([]*models.Config, error)
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
	if result == nil {
		return nil, ErrConfigNotFound
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
		return nil, fmt.Errorf("err getting list of liked")
	}
	result := make([]*models.Profile, 0, len(liked))
	for _, elem := range liked {
		result = append(result, &models.Profile{Personal: elem.Personal, Criteria: elem.Criteria})
	}
	return result, nil
}

func (a *App) ListDislikedProfiles(ctx context.Context, uuid string, limit, offset int64) ([]*models.Profile, error) {
	disliked, err := a.store.ListRelated(ctx, uuid, storage.Disliked, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("err getting list of liked")
	}
	result := make([]*models.Profile, 0, len(disliked))
	for _, elem := range disliked {
		result = append(result, &models.Profile{Personal: elem.Personal, Criteria: elem.Criteria})
	}
	return result, nil
}

func (a *App) GetMatches(ctx context.Context, uuid string, count int64) ([]*models.Profile, error) {
	matches, err := a.store.ListUnrelated(ctx, uuid, count, 0)
	if err != nil {
		return nil, fmt.Errorf("err getting list of liked")
	}
	result := make([]*models.Profile, 0, len(matches))
	for _, elem := range matches {
		result = append(result, &models.Profile{Personal: elem.Personal, Criteria: elem.Criteria})
	}
	return result, nil
}
