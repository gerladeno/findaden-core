package internal

import (
	"context"
	_ "embed"
	"github.com/gerladeno/homie-core/internal/models"
	"github.com/gerladeno/homie-core/internal/storage"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"strings"
	"testing"
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
	err = store.Truncate(&models.Regions{})
	require.NoError(s.T(), err)
	for _, stmt := range strings.Split(inserts, "\n") {
		if strings.Trim(stmt, "") != "" {
			err = store.Exec(stmt)
			require.NoError(s.T(), err)
		}
	}
	s.app = NewApp(log, store)
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
			Regions:    models.Regions([]models.Region{{ID: 1}, {ID: 3}}),
			PriceRange: models.NewRange(20000, 45000),
			Gender:     0,
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
	_, err = s.app.GetConfig(context.Background(), uuid+"d")
	require.ErrorIs(s.T(), err, ErrConfigNotFound)
}

func TestLogicSuite(t *testing.T) {
	suite.Run(t, new(LogicSuite))
}
