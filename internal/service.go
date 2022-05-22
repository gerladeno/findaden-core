package internal

import (
	"context"
	"errors"
	"fmt"

	"github.com/gerladeno/homie-core/pkg/chat"

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
	GetProfiles(ctx context.Context, uuids []string) ([]*models.Profile, error)
}

type Chat interface {
	GetDialog(ctx context.Context, client, target string) *chat.Hub
	GetAllChats(ctx context.Context, uuid string) ([]string, error)
}

type App struct {
	log        *logrus.Entry
	store      Storage
	chatServer Chat
}

func NewApp(log *logrus.Logger, store Storage, chatServer Chat) *App {
	return &App{
		log:        log.WithField("module", "app"),
		store:      store,
		chatServer: chatServer,
	}
}

func (a *App) GetDialog(ctx context.Context, client, target string) *chat.Hub {
	return a.chatServer.GetDialog(ctx, client, target)
}

func (a *App) GetAllChats(ctx context.Context, uuid string) ([]*models.Profile, error) {
	uuids, err := a.chatServer.GetAllChats(ctx, uuid)
	if err != nil {
		return nil, fmt.Errorf("err getting list of uuids client chatted with: %w", err)
	}
	profiles, err := a.store.GetProfiles(ctx, uuids)
	if err != nil {
		return nil, fmt.Errorf("err getting list of profiles client chatted with: %w", err)
	}
	return profiles, nil
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
		return fmt.Errorf("err adding relation: %w", err)
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
		return nil, fmt.Errorf("err getting list of matches: %w", err)
	}
	return matches, nil
}

func (a *App) GetProfiles(ctx context.Context, uuids []string) ([]*models.Profile, error) {
	return a.store.GetProfiles(ctx, uuids)
}
