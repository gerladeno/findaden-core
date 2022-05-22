package storage

import "github.com/gerladeno/homie-core/internal/models"

type SearchCriteria struct {
	UUID      string   `db:"uuid"`
	Regions   []int64  `db:"regions"`
	PriceFrom *float64 `db:"price_from"`
	PriceTo   *float64 `db:"price_to"`
	Gender    int8     `db:"gender"`
	AgeFrom   *float64 `db:"age_from"`
	AgeTo     *float64 `db:"age_to"`
}

func DBCriteria2Model(dbCriteria *SearchCriteria, criteria *models.SearchCriteria) {
	if dbCriteria == nil {
		criteria = nil
		return
	}
	criteria.UUID = dbCriteria.UUID
	criteria.Regions = dbCriteria.Regions

	criteria.PriceRange = models.Range{From: dbCriteria.PriceFrom, To: dbCriteria.PriceTo}
	criteria.Gender = models.Gender(dbCriteria.Gender)
	criteria.AgeRange = models.Range{From: dbCriteria.AgeFrom, To: dbCriteria.AgeTo}
}

type Profile struct {
	UUID           string   `db:"uuid"`
	Regions        []int64  `db:"regions"`
	PriceFrom      *float64 `db:"price_from"`
	PriceTo        *float64 `db:"price_to"`
	CriteriaGender int8     `db:"criteria_gender"`
	AgeFrom        *float64 `db:"age_from"`
	AgeTo          *float64 `db:"age_to"`
	Username       string   `db:"username"`
	AvatarLink     string   `db:"avatar_link"`
	PersonalGender int8     `db:"personal_gender"`
	Age            int8     `db:"age"`
}

func DBProfile2Profile(profile *Profile) *models.Profile {
	if profile == nil {
		return nil
	}
	p := models.Profile{
		UUID: profile.UUID,
		Personal: &models.Personal{
			UUID:       profile.UUID,
			Username:   profile.Username,
			AvatarLink: profile.AvatarLink,
			Gender:     models.Gender(profile.PersonalGender),
			Age:        profile.Age,
		},
		Criteria: &models.SearchCriteria{
			UUID:       profile.UUID,
			Regions:    profile.Regions,
			PriceRange: models.Range{From: profile.PriceFrom, To: profile.PriceTo},
			Gender:     models.Gender(profile.CriteriaGender),
			AgeRange:   models.Range{From: profile.AgeFrom, To: profile.AgeTo},
		},
	}
	return &p
}
