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
	); err != nil {
		return fmt.Errorf("err migrating postgres")
	}
	return nil
}

func (s *Storage) SaveConfig(ctx context.Context, config *models.Config) error {
	if config == nil {
		return nil
	}
	if s.db.WithContext(ctx).Model(&config).Where("uuid = ?", config.UUID).Updates(&config).RowsAffected == 0 {
		s.db.WithContext(ctx).Create(&config)
	}
	return s.db.Error
}

func (s *Storage) GetConfig(ctx context.Context, uuid string) (*models.Config, error) {
	var cfg models.Config
	s.db.WithContext(ctx).Model(&cfg).Where("uuid = ?", uuid).First(&cfg)
	return &cfg, s.db.Error
}

func (s *Storage) GetRegions(ctx context.Context) ([]*models.Region, error) {
	var regions []*models.Region
	s.db.WithContext(ctx).Model(&models.Region{}).Find(&regions)
	return regions, s.db.Error
}
