package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"gorm.io/gorm"
)

type Gender int8

func (g Gender) Value() (driver.Value, error) {
	return int8(g), nil
}

func (g *Gender) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	v, ok := src.(int8)
	if !ok {
		return errors.New("err scanning gender")
	}
	*g = Gender(v)
	return nil
}

const (
	Any Gender = iota
	Male
	Female
)

type Config struct {
	gorm.Model
	UUID       string         `json:"uuid" gorm:"uniqueIndex"`
	Personal   Personal       `json:"personal" gorm:"foreignkey:UUID;references:UUID"`
	Criteria   SearchCriteria `json:"criteria" gorm:"foreignkey:UUID;references:UUID"`
	Appearance Appearance     `json:"appearance" gorm:"foreignkey:UUID;references:UUID"`
}

type Profile struct {
	Personal Personal       `json:"personal"`
	Criteria SearchCriteria `json:"criteria"`
}

type Personal struct {
	gorm.Model
	UUID       string `json:"uuid"`
	Username   string `json:"username"`
	AvatarLink string `json:"avatar_link"`
	Gender     Gender `json:"gender"`
	Age        int8   `json:"age"`
}

type Relation struct {
	gorm.Model
	UUID     string
	Target   string
	Relation int8
}

type Appearance struct {
	gorm.Model
	UUID  string `json:"uuid"`
	Theme int64  `json:"theme"`
}

type SearchCriteria struct {
	gorm.Model
	UUID       string  `json:"uuid"`
	Regions    Regions `json:"regions,omitempty" gorm:"many2many:criteria_region;"`
	PriceRange Range   `json:"price_range" gorm:"embedded"`
	Gender     Gender  `json:"gender"`
	AgeRange   Range   `json:"age_range" gorm:"embedded"`
}

type Regions []Region

func (r *Regions) UnmarshalJSON(b []byte) error {
	var dest []int64
	if err := json.Unmarshal(b, &dest); err != nil {
		return err
	}
	result := make([]Region, 0, len(dest))
	for _, elem := range dest {
		result = append(result, Region{ID: elem})
	}
	*r = result
	return nil
}

func (r Regions) MarshalJSON() ([]byte, error) {
	return json.Marshal(r)
}

type Region struct {
	gorm.Model
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Range struct {
	From *float64 `json:"from,omitempty"`
	To   *float64 `json:"to,omitempty"`
}
