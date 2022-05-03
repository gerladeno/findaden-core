package models

import (
	"database/sql/driver"
	"fmt"
)

type Gender int8

func (g *Gender) Value() (driver.Value, error) {
	return int8(*g), nil
}

func (g *Gender) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	v, err := driver.Int32.ConvertValue(src)
	if err != nil {
		return fmt.Errorf("err scanning gender: %w", err)
	}
	val, ok := v.(int64)
	if !ok {
		return fmt.Errorf("err scanning gender")
	}
	*g = Gender(val)
	return nil
}

const (
	Any Gender = iota
	Male
	Female
)

type Config struct {
	UUID       string         `json:"uuid" gorm:"primarykey"`
	Personal   Personal       `json:"personal" gorm:"foreignkey:UUID;references:UUID"`
	Criteria   SearchCriteria `json:"criteria" gorm:"foreignkey:UUID;references:UUID"`
	Appearance Appearance     `json:"appearance" gorm:"foreignkey:UUID;references:UUID"`
}

func (c *Config) SetUUID(uuid string) {
	c.UUID = uuid
	c.Personal.UUID = uuid
	c.Criteria.UUID = uuid
	c.Appearance.UUID = uuid
}

type Profile struct {
	Personal Personal       `json:"personal"`
	Criteria SearchCriteria `json:"criteria"`
}

type Personal struct {
	UUID       string `json:"uuid" gorm:"primarykey"`
	Username   string `json:"username"`
	AvatarLink string `json:"avatar_link"`
	Gender     Gender `json:"gender"`
	Age        int8   `json:"age"`
}

type Relation struct {
	UUID     string `gorm:"primaryKey"`
	Target   string `gorm:"primaryKey"`
	Relation int8
}

type Appearance struct {
	UUID  string `json:"uuid" gorm:"primarykey"`
	Theme int64  `json:"theme"`
}

type SearchCriteria struct {
	UUID       string   `json:"uuid" gorm:"primarykey"`
	Regions    []Region `json:"regions" gorm:"many2many:uuid_regions"`
	PriceRange Range    `json:"price_range" gorm:"embedded;embeddedPrefix:price_"`
	Gender     Gender   `json:"gender"`
	AgeRange   Range    `json:"age_range" gorm:"embedded;embeddedPrefix:age_"`
}

type Region struct {
	ID          int64  `json:"id" gorm:"primarykey"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Range struct {
	From *float64 `json:"from,omitempty"`
	To   *float64 `json:"to,omitempty"`
}

func NewRange(from, to float64) Range {
	r := Range{}
	if from != 0 {
		r.From = &from
	}
	if to != 0 {
		r.To = &to
	}
	return r
}
