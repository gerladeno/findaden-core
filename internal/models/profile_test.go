package models

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConfig(t *testing.T) {
	t.Run("generate json", func(t *testing.T) {
		conf := Config{
			UUID: "797bcfb5-ca07-11ec-a6c3-049226c2eb3c",
			Personal: &Personal{
				UUID:       "797bcfb5-ca07-11ec-a6c3-049226c2eb3c",
				Username:   "chuvak",
				AvatarLink: "jopa.ru",
				Gender:     1,
				Age:        26,
			},
			Criteria: &SearchCriteria{
				UUID:       "797bcfb5-ca07-11ec-a6c3-049226c2eb3c",
				Regions:    []int64{1, 2, 3},
				PriceRange: NewRange(35000, 700000),
				Gender:     0,
				AgeRange:   NewRange(20, 35),
			},
			Settings: &Settings{
				UUID:  "797bcfb5-ca07-11ec-a6c3-049226c2eb3c",
				Theme: 0,
			},
		}
		b, err := json.Marshal(conf)
		require.NoError(t, err)
		fmt.Println(string(b))
	})
}

func TestConfigMinimal(t *testing.T) {
	t.Run("generate json", func(t *testing.T) {
		conf := Config{
			Personal: &Personal{
				Username:   "chuvak",
				AvatarLink: "jopa.ru",
				Gender:     1,
				Age:        26,
			},
			Criteria: &SearchCriteria{
				Regions:    []int64{1, 2, 3},
				PriceRange: NewRange(35000, 700000),
				Gender:     0,
				AgeRange:   NewRange(20, 35),
			},
			Settings: &Settings{
				Theme: 0,
			},
		}
		b, err := json.Marshal(conf)
		require.NoError(t, err)
		fmt.Println(string(b))
	})
}
