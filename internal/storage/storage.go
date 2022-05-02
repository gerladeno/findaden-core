package storage

import (
	"context"
	"fmt"

	"github.com/gerladeno/homie-core/internal/models"
	"github.com/gerladeno/homie-core/pkg/metrics"
	"github.com/jackc/pgx/v4"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Relation int8

const (
	Liked Relation = iota
	SuperLiked
	Disliked
	Neither
)

type Storage struct {
	log     *logrus.Entry
	db      *gorm.DB
	metrics *metrics.DBClient
}

func New(log *logrus.Logger, dsn string) (*Storage, error) {
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	s := Storage{log: log.WithField("module", "storage")}
	s.db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("err connecting to postgres")
	}
	fn := func() float64 {
		return 1.0
	}
	s.metrics = metrics.NewDBClient(config.Database, config.Host, fmt.Sprintf("%d", config.Port), fn).AutoRegister()
	return &s, nil
}

func (s *Storage) Migrate() error {
	if err := s.db.AutoMigrate(
		&models.Config{},
		&models.Personal{},
		&models.Appearance{},
		&models.SearchCriteria{},
		&models.Region{},
		&models.Relation{},
	); err != nil {
		return fmt.Errorf("err migrating postgres")
	}
	return nil
}

func (s *Storage) SaveConfig(ctx context.Context, config *models.Config) error {
	if config == nil {
		return nil
	}
	if s.db.WithContext(ctx).Model(config).Where("uuid = ?", config.UUID).Updates(config).RowsAffected == 0 {
		s.db.WithContext(ctx).Create(config)
	}
	return s.db.Error
}

func (s *Storage) GetConfig(ctx context.Context, uuid string) (*models.Config, error) {
	var cfg models.Config
	s.db.WithContext(ctx).Model(&cfg).Where("uuid = ?", uuid).First(&cfg)
	if s.db.RowsAffected == 0 {
		return nil, nil //nolint:nilnil
	}
	return &cfg, s.db.Error
}

func (s *Storage) GetRegions(ctx context.Context) ([]*models.Region, error) {
	var regions []*models.Region
	s.db.WithContext(ctx).Model(&models.Region{}).Find(&regions)
	return regions, s.db.Error
}

func (s *Storage) UpsertRelation(ctx context.Context, relation *models.Relation) error {
	if relation == nil {
		return nil
	}
	if s.db.WithContext(ctx).Model(relation).
		Where("uuid = ?", relation.UUID).
		Where("target_uuid = ?", relation.Target).
		Updates(relation).RowsAffected == 0 {
		s.db.WithContext(ctx).Create(relation)
	}
	return s.db.Error
}

func (s *Storage) ListRelated(ctx context.Context, uuid string, relation Relation, limit, offset int64) ([]*models.Config, error) { //nolint:lll
	var uuids []string
	if err := s.db.WithContext(ctx).
		Select("target_uuid").
		Find(uuids, "uuid = ?", uuid, "relation = ?", relation).Error; err != nil {
		return nil, err
	}
	var result []*models.Config
	s.db.WithContext(ctx).Limit(int(limit)).Offset(int(offset)).Find(&result, "uuid IN (?)", uuids)
	return result, s.db.Error
}

func (s *Storage) ListUnrelated(ctx context.Context, uuid string, limit, offset int64) ([]*models.Config, error) {
	var uuids []string
	if err := s.db.WithContext(ctx).
		Select("target_uuid").
		Find(uuids, "uuid = ?", uuid).Error; err != nil {
		return nil, err
	}
	var result []*models.Config
	s.db.WithContext(ctx).Limit(int(limit)).Offset(int(offset)).Find(&result, "uuid NOT IN (?)", uuids)
	return result, s.db.Error
}
