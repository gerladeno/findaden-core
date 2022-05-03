package storage

import (
	"context"
	"fmt"
	"github.com/gerladeno/homie-core/internal/models"
	"github.com/gerladeno/homie-core/pkg/metrics"
	"github.com/jackc/pgx/v4"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
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

func NewSQLiteStore(log *logrus.Entry, filename string) (*Storage, error) {
	db, err := gorm.Open(sqlite.Open(filename), &gorm.Config{
		//Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("err connecting to postgres")
	}
	fn := func() float64 {
		return 1.0
	}
	return &Storage{
		log:     log,
		db:      db,
		metrics: metrics.NewDBClient("", filename, "", fn).AutoRegister(),
	}, nil
}

func (s *Storage) Exec(query string) error {
	return s.db.Exec(query).Error
}

func (s *Storage) Truncate(tables ...interface{}) error {
	var err error
	db := s.db.Session(&gorm.Session{AllowGlobalUpdate: true})
	for _, table := range tables {
		if err = db.Unscoped().Delete(table).Error; err != nil {
			return err
		}
	}
	return nil
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
	s.db.Session(&gorm.Session{FullSaveAssociations: true, Context: ctx}).Model(config).
		Omit("Criteria.Regions.*").
		Save(config)
	return s.db.WithContext(ctx).
		Model(&config.Criteria).
		Association("Regions").
		Replace(&config.Criteria.Regions)
}

func (s *Storage) GetConfig(ctx context.Context, uuid string) (*models.Config, error) {
	var cfg models.Config
	if err := s.db.WithContext(ctx).
		Model(&cfg).
		Preload("Personal").
		Preload("Criteria").
		Preload("Criteria.Regions").
		Preload("Appearance").
		First(&cfg, "uuid = ?", uuid).Error; err != nil {
		return nil, err
	}
	return &cfg, nil
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
		Updates(relation).RowsAffected == 0 {
		s.db.Create(relation)
	}
	return s.db.Error
}

func (s *Storage) ListRelated(ctx context.Context, uuid string, relation Relation, limit, offset int64) ([]*models.Config, error) { //nolint:lll
	var uuids []string
	if err := s.db.WithContext(ctx).
		Model(&models.Relation{}).
		Select("target").
		Find(&uuids, "uuid = ?", uuid, "relation = ?", relation).Error; err != nil {
		return nil, err
	}
	var result []*models.Config
	err := s.db.Limit(int(limit)).
		Offset(int(offset)).
		Preload("Personal").
		Preload("Criteria").
		Preload("Criteria.Regions").
		Preload("Appearance").
		Find(&result, "uuid IN (?)", uuids).Error
	return result, err
}

func (s *Storage) ListMatches(ctx context.Context, uuid string, count int64) ([]*models.Config, error) {
	var config models.Config
	if err := s.db.WithContext(ctx).
		Model(&config).
		Preload("Personal").
		Preload("Criteria").
		Preload("Criteria.Regions").
		First(&config, "uuid = ?", uuid).Error; err != nil {
		return nil, err
	}
	matchingUUIDS, err := s.getCriteria(ctx, &config)
	if err != nil {
		return nil, err
	}
	var metUUIDs []string
	if err = s.db.WithContext(ctx).Model(&models.Relation{}).Select("target").
		Find(&metUUIDs, "uuid = ?", uuid).Error; err != nil {
		return nil, err
	}
	var result []*models.Config
	err = s.db.WithContext(ctx).
		Limit(int(count)).
		Preload("Personal").
		Preload("Criteria").
		Preload("Criteria.Regions").
		Preload("Appearance").
		Not(map[string]interface{}{"uuid": metUUIDs}).
		Where(map[string]interface{}{"uuid": matchingUUIDS}).
		Find(&result).Error
	return result, err
}

const queryJoinRegions = `
select distinct others.search_criteria_uuid
from (select *
      from uuid_regions
      where search_criteria_uuid = ?) own
         join (select *
               from uuid_regions
               where search_criteria_uuid != ?) others
              on own.region_id = others.region_id
`

func (s *Storage) getCriteria(ctx context.Context, config *models.Config) ([]string, error) {
	personal := s.db.WithContext(ctx).Model(&models.Personal{}).Select("uuid")
	if config.Criteria.AgeRange.From != nil {
		personal = personal.Where("age >= ?", *config.Criteria.AgeRange.From)
	}
	if config.Criteria.AgeRange.To != nil {
		personal = personal.Where("age <= ?", *config.Criteria.AgeRange.To)
	}
	if config.Criteria.Gender != models.Any {
		personal = personal.Where("gender = ?", config.Criteria.Gender)
	}
	var uuidsByPersonal []string
	if err := personal.Find(&uuidsByPersonal).Error; err != nil {
		return nil, err
	}
	if len(uuidsByPersonal) == 0 {
		return nil, nil //nolint:nilnil
	}
	criteria := s.db.WithContext(ctx).Model(&models.SearchCriteria{}).Select("uuid")
	if config.Criteria.PriceRange.From != nil {
		criteria = criteria.Where("to >= ?", *config.Criteria.PriceRange.From)
	}
	if config.Criteria.PriceRange.To != nil {
		criteria = criteria.Where("from <= ?", *config.Criteria.PriceRange.To)
	}
	if len(config.Criteria.Regions) > 0 {
		criteria = criteria.Raw(queryJoinRegions, config.UUID, config.UUID)
	}
	var uuidsByCriteria []string
	if err := criteria.Find(&uuidsByCriteria).Error; err != nil {
		return nil, err
	}
	if len(uuidsByCriteria) == 0 {
		return nil, nil //nolint:nilnil
	}
	return findMatchingStrings(uuidsByPersonal, uuidsByCriteria), nil
}

func findMatchingStrings(s1 []string, s2 []string) []string {
	var result []string
	m := make(map[string]struct{})
	for _, s := range s1 {
		m[s] = struct{}{}
	}
	for _, s := range s2 {
		if _, ok := m[s]; ok {
			result = append(result, s)
		}
	}
	return result
}
