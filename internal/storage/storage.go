package storage

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/gerladeno/homie-core/pkg/common"
	"github.com/jackc/pgx/v4/pgxpool"
	migrate "github.com/rubenv/sql-migrate"
	"strings"
	"time"

	"github.com/gerladeno/homie-core/internal/models"
	"github.com/gerladeno/homie-core/pkg/metrics"
	"github.com/jackc/pgx/v4"
	"github.com/sirupsen/logrus"
)

//go:embed migrations
var migrations embed.FS

type Relation int8

const (
	Liked Relation = iota
	SuperLiked
	Disliked
	Neither
)

type Storage struct {
	log     *logrus.Entry
	db      *pgxpool.Pool
	dsn     string
	metrics *metrics.DBClient
}

func New(ctx context.Context, log *logrus.Logger, dsn string) (*Storage, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	s := Storage{log: log.WithField("module", "storage"), dsn: dsn}
	s.db, err = pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("err connecting to postgres")
	}
	fn := func() float64 {
		return 1.0
	}
	s.metrics = metrics.NewDBClient(config.ConnConfig.Database, config.ConnConfig.Host, fmt.Sprintf("%d", config.ConnConfig.Port), fn).AutoRegister()
	return &s, nil
}

func (s *Storage) Exec(ctx context.Context, query string) error {
	_, err := s.db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("err executing %s: %w", query, err)
	}
	return nil
}

func (s *Storage) Truncate(ctx context.Context, tables ...string) error {
	var err error
	for _, table := range tables {
		if err = s.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)); err != nil {
			return fmt.Errorf("err truncating %s: %w", table, err)
		}
	}
	return nil
}

func (s *Storage) Migrate() error {
	conn, err := sql.Open("pgx", s.dsn)
	if err != nil {
		return err
	}
	defer func() {
		if err = conn.Close(); err != nil {
			s.log.Error("err closing migration connection")
		}
	}()
	assetDir := func() func(string) ([]string, error) {
		return func(path string) ([]string, error) {
			dirEntry, er := migrations.ReadDir(path)
			if er != nil {
				return nil, er
			}
			entries := make([]string, 0)
			for _, e := range dirEntry {
				entries = append(entries, e.Name())
			}

			return entries, nil
		}
	}()
	asset := migrate.AssetMigrationSource{
		Asset:    migrations.ReadFile,
		AssetDir: assetDir,
		Dir:      "migrations",
	}
	_, err = migrate.Exec(conn, "postgres", asset, migrate.Up)
	return err
}

func (s *Storage) SaveConfig(ctx context.Context, config *models.Config) error {
	if config == nil {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return fmt.Errorf("err saving config: %w", err)
	}
	defer func() {
		if err = tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			s.log.Warnf("err rolling back tx during saving config: %v", err)
		}
	}()
	if err = s.upsertConfig(tx, ctx, config); err != nil {
		return fmt.Errorf("err saving config: %w", err)
	}
	if err = s.upsertSettings(tx, ctx, config.Settings); err != nil {
		return fmt.Errorf("err saving config: %w", err)
	}
	if err = s.upsertPersonal(tx, ctx, config.Personal); err != nil {
		return fmt.Errorf("err saving config: %w", err)
	}
	if err = s.upsertCriteria(tx, ctx, config.Criteria); err != nil {
		return fmt.Errorf("err saving config: %w", err)
	}
	if config.Criteria != nil && config.Criteria.Regions != nil {
		if err = s.updateCriteriaRegions(tx, ctx, config.Criteria.UUID, config.Criteria.Regions); err != nil {
			return fmt.Errorf("err saving config: %w", err)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("err commiting save config transaction: %w", err)
	}
	return nil
}

func (s *Storage) upsertConfig(tx pgx.Tx, ctx context.Context, config *models.Config) error {
	query := `
INSERT INTO config (uuid, created, updated)
VALUES ($1, $2, $3)
ON CONFLICT (uuid) DO UPDATE SET updated = EXCLUDED.updated
`
	t := time.Now()
	res, err := tx.Exec(ctx, query, config.UUID, t, t)
	if err != nil {
		return fmt.Errorf("err inserting config for %s: %w", config.UUID, err)
	}
	if res.RowsAffected() == 0 {
		return errors.New("err no rows affected while upserting config")
	}
	return nil
}

func (s *Storage) upsertSettings(tx pgx.Tx, ctx context.Context, settings *models.Settings) error {
	if settings == nil {
		return nil
	}
	query := `
INSERT INTO settings (uuid, theme)
VALUES ($1, $2)
ON CONFLICT (uuid) DO UPDATE SET theme = excluded.theme
`
	res, err := tx.Exec(ctx, query, settings.UUID, settings.Theme)
	if err != nil {
		return fmt.Errorf("err inserting settings for %s: %w", settings.UUID, err)
	}
	if res.RowsAffected() == 0 {
		return errors.New("err no rows affected while upserting settings")
	}
	return nil
}

func (s *Storage) upsertPersonal(tx pgx.Tx, ctx context.Context, personal *models.Personal) error {
	if personal == nil {
		return nil
	}
	query := `
INSERT INTO personal (uuid, username, avatar_link, gender, age)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (uuid) DO UPDATE SET username = excluded.username,
								 avatar_link = excluded.avatar_link,
								 gender = excluded.gender,
								 age = excluded.age
`
	res, err := tx.Exec(ctx, query, personal.UUID, personal.Username, personal.AvatarLink, personal.Gender, personal.Age)
	if err != nil {
		return fmt.Errorf("err inserting personal for %s: %w", personal.UUID, err)
	}
	if res.RowsAffected() == 0 {
		return errors.New("err no rows affected while upserting personal")
	}
	return nil
}

func (s *Storage) upsertCriteria(tx pgx.Tx, ctx context.Context, criteria *models.SearchCriteria) error {
	if criteria == nil {
		return nil
	}
	query := `
INSERT INTO search_criteria (uuid, price_from, price_to, gender, age_from, age_to)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (uuid) DO UPDATE SET price_from = excluded.price_from,
								 price_to = excluded.price_to,
								 gender = excluded.gender,
								 age_from = excluded.age_from,
								 age_to = excluded.age_to
`
	res, err := tx.Exec(ctx, query,
		criteria.UUID,
		criteria.PriceRange.From,
		criteria.PriceRange.To,
		criteria.Gender,
		criteria.AgeRange.From,
		criteria.AgeRange.To,
	)
	if err != nil {
		return fmt.Errorf("err inserting criteria for %s: %w", criteria.UUID, err)
	}
	if res.RowsAffected() == 0 {
		return errors.New("err no rows affected while upserting criteria")
	}
	return nil
}

func (s *Storage) updateCriteriaRegions(tx pgx.Tx, ctx context.Context, uuid string, regions []int64) error {
	query := `
DELETE FROM uuid_regions
WHERE uuid = $1
`
	_, err := tx.Exec(ctx, query, uuid)
	if err != nil {
		return fmt.Errorf("err deleting old regions for %s: %w", uuid, err)
	}
	var rows [][]interface{}
	for _, region := range regions {
		rows = append(rows, []interface{}{uuid, region})
	}
	_, err = tx.CopyFrom(
		ctx, pgx.Identifier{"uuid_regions"},
		[]string{"uuid", "region_id"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("err inserting criteria regions: %w", err)
	}
	return nil
}

func (s *Storage) GetConfig(ctx context.Context, uuid string) (*models.Config, error) {
	var cfg models.Config
	cfg.UUID = uuid
	if err := s.getConfig(ctx, uuid); err != nil {
		return nil, err
	}
	settings := models.Settings{}
	err := s.getSettings(ctx, uuid, &settings)
	switch {
	case err == nil:
		cfg.Settings = &settings
	case errors.Is(err, pgx.ErrNoRows):
	default:
		return nil, fmt.Errorf("err getting config for %s: %w", uuid, err)
	}
	personal := models.Personal{}
	err = s.getPersonal(ctx, uuid, &personal)
	switch {
	case err == nil:
		cfg.Personal = &personal
	case errors.Is(err, pgx.ErrNoRows):
	default:
		return nil, fmt.Errorf("err getting config for %s: %w", uuid, err)
	}
	criteria := models.SearchCriteria{}
	err = s.getSearchCriteria(ctx, uuid, &criteria)
	switch {
	case err == nil:
		cfg.Criteria = &criteria
	case errors.Is(err, pgx.ErrNoRows):
	default:
		return nil, fmt.Errorf("err getting config for %s: %w", uuid, err)
	}
	return &cfg, nil
}

func (s *Storage) getConfig(ctx context.Context, uuid string) error {
	row := s.db.QueryRow(ctx, `SELECT uuid FROM config WHERE uuid = $1`, uuid)
	scannedUUID := ""
	err := row.Scan(&scannedUUID)
	switch {
	case err == nil:
	case errors.Is(err, pgx.ErrNoRows):
		return common.ErrConfigNotFound
	default:
		return fmt.Errorf("err getting config for %s: %w", uuid, err)
	}
	return nil
}

func (s *Storage) getSettings(ctx context.Context, uuid string, settings *models.Settings) error {
	return pgxscan.Get(ctx, s.db, settings, `SELECT uuid, theme FROM settings WHERE uuid = $1`, uuid)
}

func (s *Storage) getPersonal(ctx context.Context, uuid string, personal *models.Personal) error {
	return pgxscan.Get(ctx, s.db, personal,
		`SELECT uuid, username, avatar_link, gender, age FROM personal WHERE uuid = $1`, uuid)
}

func (s *Storage) getSearchCriteria(ctx context.Context, uuid string, criteria *models.SearchCriteria) error {
	dbCriteria := SearchCriteria{}
	err := pgxscan.Get(ctx, s.db, &dbCriteria,
		`
SELECT uuid,
       (select array (select distinct region_id from uuid_regions where uuid = $1)) as regions,
       price_from,
       price_to,
       gender,
       age_from,
       age_to
FROM search_criteria WHERE uuid = $2`, uuid, uuid)
	if err != nil {
		return err
	}
	DBCriteria2Model(&dbCriteria, criteria)
	return nil
}

func (s *Storage) GetRegions(ctx context.Context) ([]*models.Region, error) {
	var regions []*models.Region
	err := pgxscan.Select(ctx, s.db, &regions, "SELECT * FROM regions")
	if err != nil {
		return nil, fmt.Errorf("err getting regions: %w", err)
	}
	return regions, nil
}

func (s *Storage) UpsertRelation(ctx context.Context, relation *models.Relation) error {
	if relation == nil {
		return nil
	}
	query := `
INSERT INTO relations (uuid, target, relation)
VALUES ($1, $2, $3)
ON CONFLICT (uuid, target) DO UPDATE SET relation = excluded.relation
`
	res, err := s.db.Exec(ctx, query, relation.UUID, relation.Target, relation.Relation)
	if err != nil {
		return fmt.Errorf("err inserting relation for %s and %s: %w", relation.UUID, relation.Target, err)
	}
	if res.RowsAffected() == 0 {
		return errors.New("err no rows affected while upserting relation")
	}
	return nil
}

func (s *Storage) ListRelated(ctx context.Context, uuid string, relation Relation, limit, offset int64) ([]*models.Profile, error) { //nolint:lll
	var uuids []string
	query := `SELECT target FROM relations WHERE uuid = $1 AND relation = $2`
	if limit != 0 {
		query += fmt.Sprintf("\nLIMIT %d OFFSET %d", limit, offset)
	}
	err := pgxscan.Select(ctx, s.db, &uuids, query, uuid, relation)
	switch {
	case err == nil:
	case errors.Is(err, pgx.ErrNoRows):
		return nil, nil
	default:
		return nil, fmt.Errorf("err selecting relations: %w", err)
	}
	var result []*models.Profile
	err = s.getProfiles(ctx, &result, uuids)
	if err != nil {
		return nil, fmt.Errorf("err selecting related profiles: %w", err)
	}
	return result, nil
}

func (s *Storage) getProfiles(ctx context.Context, profiles *[]*models.Profile, uuids []string) error {
	if uuids == nil {
		return nil
	}
	var dbProfiles []Profile
	quotedUUIDs := strings.Join(wrapQuoted(uuids), ",")
	query := fmt.Sprintf(`
SELECT criteria.uuid   AS uuid,
       regions,
       price_from,
       price_to,
       criteria.gender AS criteria_gender,
       age_from,
       age_to,
       username,
       avatar_link,
       personal.gender AS personal_gender,
       age
FROM (SELECT search_criteria.uuid,
       (select array (select distinct region_id from uuid_regions where uuid IN (%[1]s))) as regions,
       price_from,
       price_to,
       gender,
       age_from,
       age_to
FROM search_criteria LEFT JOIN uuid_regions ON search_criteria.uuid = uuid_regions.uuid
WHERE search_criteria.uuid IN (%[1]s)) AS criteria
JOIN (SELECT uuid, username, avatar_link, gender, age FROM personal WHERE uuid IN (%[1]s)
) AS personal
ON personal.uuid = criteria.uuid`, quotedUUIDs)
	err := pgxscan.Select(ctx, s.db, &dbProfiles, query)
	switch {
	case err == nil:
		for _, profile := range dbProfiles {
			*profiles = append(*profiles, DBProfile2Profile(&profile))
		}
	case errors.Is(err, pgx.ErrNoRows):
	default:
		return fmt.Errorf("err getting profiles: %w", err)
	}
	return nil
}

func (s *Storage) ListMatches(ctx context.Context, uuid string, count int64) ([]*models.Profile, error) {
	config, err := s.GetConfig(ctx, uuid)
	if err != nil {
		return nil, fmt.Errorf("err getting matches: %w", err)
	}
	var uuids []string
	err = pgxscan.Select(ctx, s.db, &uuids,
		`
SELECT DISTINCT uuid
FROM uuid_regions
WHERE region_id IN (SELECT region_id FROM uuid_regions WHERE uuid = $1);
`, uuid)
	switch {
	case err == nil:
	case errors.Is(err, pgx.ErrNoRows):
		return nil, nil
	default:
		return nil, fmt.Errorf("err selecting matches by region: %w", err)
	}
	matchingUUIDS, err := s.getMatchingUUIDs(ctx, config)
	if err != nil {
		return nil, err
	}
	var metUUIDs []string
	err = pgxscan.Select(ctx, s.db, &metUUIDs,
		`
SELECT DISTINCT target
FROM relations
WHERE uuid = $1;
`, uuid)
	var result []*models.Profile
	err = s.getProfiles(ctx, &result, matchingUUIDS)
	if err != nil {
		return nil, fmt.Errorf("err getting matches for %s: %w", uuid, err)
	}
	return result, nil
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

func (s *Storage) getMatchingUUIDs(ctx context.Context, config *models.Config) ([]string, error) {
	query := `
SELECT uuid FROM personal
WHERE 1=1
`
	if config.Criteria.AgeRange.From != nil {
		query += fmt.Sprintf("AND age >= %f", *config.Criteria.AgeRange.From)
	}
	if config.Criteria.AgeRange.To != nil {
		query += fmt.Sprintf("AND age >= %f", *config.Criteria.AgeRange.To)
	}
	if config.Criteria.Gender != models.Any {
		query += fmt.Sprintf("AND gender >= %d", config.Criteria.Gender)
	}
	var uuidsByPersonal []string
	if err := pgxscan.Select(ctx, s.db, &uuidsByPersonal, query); err != nil {
		return nil, err
	}
	if len(uuidsByPersonal) == 0 {
		return nil, nil
	}
	query = `
SELECT uuid FROM search_criteria
WHERE 1=1
`
	if config.Criteria.PriceRange.From != nil {
		query += fmt.Sprintf("AND price_to >= %f", *config.Criteria.PriceRange.From)
	}
	//if config.Criteria.PriceRange.To != nil {
	//	criteria = criteria.Where("from <= ?", *config.Criteria.PriceRange.To)
	//}
	//if len(config.Criteria.Regions) > 0 {
	//	criteria = criteria.Raw(queryJoinRegions, config.UUID, config.UUID)
	//}
	var uuidsByCriteria []string
	//if err := criteria.Find(&uuidsByCriteria).Error; err != nil {
	//	return nil, err
	//}
	//if len(uuidsByCriteria) == 0 {
	//	return nil, nil
	//}
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

func wrapQuoted(elems []string) []string {
	result := make([]string, 0, len(elems))
	for _, elem := range elems {
		result = append(result, "'"+elem+"'")
	}
	return result
}
