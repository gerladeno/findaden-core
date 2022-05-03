package internal

import (
	"context"
	_ "embed"
	"os"
	"strings"
	"testing"

	"github.com/gerladeno/homie-core/internal/models"
	"github.com/gerladeno/homie-core/internal/storage"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type LogicSuite struct {
	app *App
	suite.Suite
}

const filename = "test.db"

//go:embed storage/data/insert_data.sql
var inserts string

func (s *LogicSuite) SetupSuite() {
	log := logrus.New()
	store, err := storage.NewSQLiteStore(log.WithField("test", "db"), filename)
	require.NoError(s.T(), err)
	err = store.Migrate()
	require.NoError(s.T(), err)
	err = store.Truncate(
		&models.Region{},
	)
	require.NoError(s.T(), err)
	for _, stmt := range strings.Split(inserts, "\n") {
		if strings.Trim(stmt, "") != "" {
			err = store.Exec(stmt)
			require.NoError(s.T(), err)
		}
	}
	s.app = NewApp(log, store)
}

func (s *LogicSuite) SetupTest() {
	err := s.app.store.(*storage.Storage).Truncate(
		&models.Config{},
		&models.SearchCriteria{},
		&models.Personal{},
		&models.Appearance{},
		&models.Relation{},
	)
	require.NoError(s.T(), err)
}

func (s *LogicSuite) TearDownSuite() {
	err := os.Remove(filename)
	require.NoError(s.T(), err)
}

func (s *LogicSuite) TestSaveGetConfig() {
	uuid := "797bcfb5-ca07-11ec-a6c3-049226c2eb3c"
	cfg := models.Config{
		Personal: models.Personal{
			Username:   "bober",
			AvatarLink: "",
			Gender:     models.Male,
			Age:        19,
		},
		Criteria: models.SearchCriteria{
			Regions:    []models.Region{{ID: 1}, {ID: 3}},
			PriceRange: models.NewRange(20000, 45000),
			Gender:     models.Female,
			AgeRange:   models.NewRange(22, 0),
		},
		Appearance: models.Appearance{Theme: 12},
	}
	cfg.SetUUID(uuid)
	err := s.app.SaveConfig(context.Background(), &cfg)
	require.NoError(s.T(), err)
	cfg2, err := s.app.GetConfig(context.Background(), uuid)
	require.NoError(s.T(), err)
	require.Equal(s.T(), cfg.Personal.Username, cfg2.Personal.Username)
	require.Equal(s.T(), cfg.Appearance.Theme, cfg2.Appearance.Theme)
	require.Equal(s.T(), *cfg.Criteria.PriceRange.From, *cfg2.Criteria.PriceRange.From)
	require.Equal(s.T(), cfg.Criteria.Gender, cfg2.Criteria.Gender)
	_, err = s.app.GetConfig(context.Background(), uuid+"d")
	require.ErrorIs(s.T(), err, ErrConfigNotFound)
}

func (s *LogicSuite) TestUpdateGetConfig() {
	uuid := "797bcfb5-ca07-11ec-a6c3-049226c2eb3c"
	cfg := models.Config{
		Personal: models.Personal{
			Username:   "bober",
			AvatarLink: "",
			Gender:     models.Male,
			Age:        19,
		},
		Criteria: models.SearchCriteria{
			Regions:    []models.Region{{ID: 1}, {ID: 3}},
			PriceRange: models.NewRange(20000, 45000),
			Gender:     0,
			AgeRange:   models.NewRange(22, 0),
		},
		Appearance: models.Appearance{Theme: 12},
	}
	cfg.SetUUID(uuid)
	err := s.app.SaveConfig(context.Background(), &cfg)
	require.NoError(s.T(), err)
	cfg = models.Config{
		Personal: models.Personal{
			Username:   "gobel",
			AvatarLink: "",
			Gender:     models.Male,
			Age:        25,
		},
		Criteria: models.SearchCriteria{
			Regions:    []models.Region{{ID: 3}, {ID: 5}},
			PriceRange: models.NewRange(30000, 55000),
			Gender:     models.Female,
			AgeRange:   models.NewRange(22, 0),
		},
		Appearance: models.Appearance{Theme: 22},
	}
	cfg.SetUUID(uuid)
	err = s.app.SaveConfig(context.Background(), &cfg)
	require.NoError(s.T(), err)
	cfg.SetUUID(uuid + "d")
	err = s.app.SaveConfig(context.Background(), &cfg)
	require.NoError(s.T(), err)
	cfg2, err := s.app.GetConfig(context.Background(), uuid)
	require.NoError(s.T(), err)
	require.Equal(s.T(), cfg.Personal.Username, cfg2.Personal.Username)
	require.Equal(s.T(), cfg.Appearance.Theme, cfg2.Appearance.Theme)
	require.Equal(s.T(), *cfg.Criteria.PriceRange.From, *cfg2.Criteria.PriceRange.From)
	require.Equal(s.T(), cfg.Criteria.Regions, cfg2.Criteria.Regions)
}

func (s *LogicSuite) TestSaveViolateConstraint() {
	uuid := "797bcfb5-ca07-11ec-a6c3-049226c2eb3c"
	cfg := models.Config{
		Criteria: models.SearchCriteria{
			Regions:    []models.Region{{ID: 1}, {ID: 64}},
			PriceRange: models.NewRange(20000, 45000),
			Gender:     0,
			AgeRange:   models.NewRange(22, 0),
		},
	}
	cfg.SetUUID(uuid)
	err := s.app.SaveConfig(context.Background(), &cfg)
	require.NoError(s.T(), err)
}

func (s *LogicSuite) TestLikeGetLiked() {
	cfg := models.Config{}
	cfg.SetUUID("first")
	err := s.app.SaveConfig(context.Background(), &cfg)
	require.NoError(s.T(), err)
	cfg2 := models.Config{}
	cfg2.SetUUID("second")
	err = s.app.SaveConfig(context.Background(), &cfg2)
	require.NoError(s.T(), err)
	cfg3 := models.Config{}
	cfg3.SetUUID("third")
	err = s.app.SaveConfig(context.Background(), &cfg3)
	require.NoError(s.T(), err)

	err = s.app.Like(context.Background(), cfg.UUID, cfg2.UUID, true)
	require.NoError(s.T(), err)
	err = s.app.Like(context.Background(), cfg.UUID, cfg3.UUID, false)
	require.NoError(s.T(), err)
	liked, err := s.app.ListLikedProfiles(context.Background(), cfg.UUID, 10, 1)
	require.NoError(s.T(), err)
	require.Len(s.T(), liked, 1)
	require.Equal(s.T(), liked[0].Personal.UUID, cfg3.UUID)
	liked, err = s.app.ListLikedProfiles(context.Background(), cfg2.UUID, 10, 0)
	require.NoError(s.T(), err)
	require.Len(s.T(), liked, 0)
}

func (s *LogicSuite) TestDislikeGetDisliked() {
	cfg := models.Config{}
	cfg.SetUUID("first")
	err := s.app.SaveConfig(context.Background(), &cfg)
	require.NoError(s.T(), err)
	cfg2 := models.Config{}
	cfg2.SetUUID("second")
	err = s.app.SaveConfig(context.Background(), &cfg2)
	require.NoError(s.T(), err)
	cfg3 := models.Config{}
	cfg3.SetUUID("third")
	err = s.app.SaveConfig(context.Background(), &cfg3)
	require.NoError(s.T(), err)

	err = s.app.Dislike(context.Background(), cfg.UUID, cfg2.UUID)
	require.NoError(s.T(), err)
	err = s.app.Dislike(context.Background(), cfg.UUID, cfg3.UUID)
	require.NoError(s.T(), err)
	disliked, err := s.app.ListDislikedProfiles(context.Background(), cfg.UUID, 10, 1)
	require.NoError(s.T(), err)
	require.Len(s.T(), disliked, 1)
	require.Equal(s.T(), disliked[0].Personal.UUID, cfg3.UUID)
	disliked, err = s.app.ListDislikedProfiles(context.Background(), cfg2.UUID, 10, 0)
	require.NoError(s.T(), err)
	require.Len(s.T(), disliked, 0)
}

func (s *LogicSuite) TestGetMatchesByRegion() {
	cfg := models.Config{Criteria: models.SearchCriteria{
		Regions: []models.Region{{ID: 1}, {ID: 5}},
	}}
	cfg.SetUUID("first")
	err := s.app.SaveConfig(context.Background(), &cfg)
	require.NoError(s.T(), err)
	cfg2 := models.Config{Criteria: models.SearchCriteria{
		Regions: []models.Region{{ID: 2}, {ID: 3}},
	}}
	cfg2.SetUUID("second")
	err = s.app.SaveConfig(context.Background(), &cfg2)
	require.NoError(s.T(), err)

	matches, err := s.app.GetMatches(context.Background(), cfg.UUID, 10)
	require.NoError(s.T(), err)
	require.Len(s.T(), matches, 0)

	cfg3 := models.Config{Criteria: models.SearchCriteria{
		Regions: []models.Region{{ID: 2}, {ID: 5}},
	}}
	cfg3.SetUUID("third")
	err = s.app.SaveConfig(context.Background(), &cfg3)
	require.NoError(s.T(), err)

	matches, err = s.app.GetMatches(context.Background(), cfg.UUID, 10)
	require.NoError(s.T(), err)
	require.Len(s.T(), matches, 1)
	require.Equal(s.T(), matches[0].Personal.UUID, cfg3.UUID)

	matches, err = s.app.GetMatches(context.Background(), cfg3.UUID, 10)
	require.NoError(s.T(), err)
	require.Len(s.T(), matches, 2)
	require.Equal(s.T(), matches[0].Personal.UUID, cfg.UUID)
	require.Equal(s.T(), matches[1].Personal.UUID, cfg2.UUID)

	matches, err = s.app.GetMatches(context.Background(), cfg3.UUID, 1)
	require.NoError(s.T(), err)
	require.Len(s.T(), matches, 1)
}

func (s *LogicSuite) TestGetMatchesBySexAndAge() {
	cfg := models.Config{Criteria: models.SearchCriteria{
		Gender:   models.Male,
		AgeRange: models.NewRange(22, 30),
	}}
	cfg.SetUUID("first")
	err := s.app.SaveConfig(context.Background(), &cfg)
	require.NoError(s.T(), err)
	cfg2 := models.Config{
		Personal: models.Personal{Gender: models.Female, Age: 28},
		Criteria: models.SearchCriteria{
			Gender:   models.Any,
			AgeRange: models.NewRange(22, 30),
		},
	}
	cfg2.SetUUID("second")
	err = s.app.SaveConfig(context.Background(), &cfg2)
	require.NoError(s.T(), err)

	matches, err := s.app.GetMatches(context.Background(), cfg.UUID, 10)
	require.NoError(s.T(), err)
	require.Len(s.T(), matches, 0)

	cfg3 := models.Config{
		Personal: models.Personal{Gender: models.Male, Age: 28},
		Criteria: models.SearchCriteria{
			Gender:   models.Male,
			AgeRange: models.NewRange(22, 30),
		},
	}
	cfg3.SetUUID("first")
	err = s.app.SaveConfig(context.Background(), &cfg3)
	require.NoError(s.T(), err)

	matches, err = s.app.GetMatches(context.Background(), cfg.UUID, 10)
	require.NoError(s.T(), err)
	require.Len(s.T(), matches, 1)
	require.Equal(s.T(), matches[0].Personal.UUID, cfg3.UUID)
}

func (s LogicSuite) TestGetMatchesMatchButMet() {
	cfg := models.Config{Criteria: models.SearchCriteria{
		Gender:   models.Male,
		AgeRange: models.NewRange(22, 30),
	}}
	cfg.SetUUID("first")
	err := s.app.SaveConfig(context.Background(), &cfg)
	require.NoError(s.T(), err)
	cfg2 := models.Config{
		Personal: models.Personal{Gender: models.Male, Age: 28},
		Criteria: models.SearchCriteria{
			Gender:   models.Male,
			AgeRange: models.NewRange(22, 30),
		},
	}
	cfg2.SetUUID("second")
	err = s.app.SaveConfig(context.Background(), &cfg2)
	require.NoError(s.T(), err)

	matches, err := s.app.GetMatches(context.Background(), cfg.UUID, 10)
	require.NoError(s.T(), err)
	require.Len(s.T(), matches, 1)
	require.Equal(s.T(), matches[0].Personal.UUID, cfg2.UUID)
	err = s.app.Like(context.Background(), cfg.UUID, cfg2.UUID, true)
	require.NoError(s.T(), err)
	matches, err = s.app.GetMatches(context.Background(), cfg.UUID, 10)
	require.NoError(s.T(), err)
	require.Len(s.T(), matches, 0)
	err = s.app.Like(context.Background(), cfg.UUID, cfg2.UUID, false)
	require.NoError(s.T(), err)
	matches, err = s.app.GetMatches(context.Background(), cfg.UUID, 10)
	require.NoError(s.T(), err)
	require.Len(s.T(), matches, 0)
	err = s.app.Dislike(context.Background(), cfg.UUID, cfg2.UUID)
	require.NoError(s.T(), err)
	matches, err = s.app.GetMatches(context.Background(), cfg.UUID, 10)
	require.NoError(s.T(), err)
	require.Len(s.T(), matches, 0)
}

func TestLogicSuite(t *testing.T) {
	suite.Run(t, new(LogicSuite))
}
