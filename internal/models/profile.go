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
	UUID     string          `json:"uuid,omitempty"`
	Personal *Personal       `json:"personal,omitempty"`
	Criteria *SearchCriteria `json:"criteria,omitempty"`
	Settings *Settings       `json:"settings,omitempty"`
}

func (c *Config) SetUUID(uuid string) {
	c.UUID = uuid
	if c.Personal != nil {
		c.Personal.UUID = uuid
	}
	if c.Criteria != nil {
		c.Criteria.UUID = uuid
	}
	if c.Settings != nil {
		c.Settings.UUID = uuid
	}
}

type Profile struct {
	UUID     string          `json:"uuid,omitempty"`
	Personal *Personal       `json:"personal,omitempty"`
	Criteria *SearchCriteria `json:"criteria,omitempty"`
}

type Personal struct {
	UUID       string `json:"uuid,omitempty"`
	Username   string `json:"username"`
	AvatarLink string `json:"avatar_link"`
	Gender     Gender `json:"gender"`
	Age        int8   `json:"age"`
}

type Relation struct {
	UUID     string
	Target   string
	Relation int8
}

type Settings struct {
	UUID  string `json:"uuid,omitempty"`
	Theme int64  `json:"theme"`
}

type SearchCriteria struct {
	UUID       string  `json:"uuid,omitempty"`
	Regions    []int64 `json:"regions"`
	PriceRange Range   `json:"price_range"`
	Gender     Gender  `json:"gender"`
	AgeRange   Range   `json:"age_range"`
}

type Region struct {
	ID          int64  `json:"id"`
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
