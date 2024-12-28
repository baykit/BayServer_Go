package common

import (
	"bayserver-core/baykit/bayserver/docker"
)

type Cities struct {
	anyCity docker.City
	cities  []docker.City
}

func NewCities() Cities {
	return Cities{
		anyCity: nil,
		cities:  []docker.City{},
	}
}

func (c *Cities) Add(cty docker.City) {
	if cty.Name() == "*" {
		c.anyCity = cty

	} else {
		c.cities = append(c.cities, cty)
	}
}

func (c *Cities) FindCity(name string) docker.City {
	// Check exact match
	for _, cty := range c.cities {
		if cty.Name() == name {
			return cty
		}
	}

	return c.anyCity
}

func (c *Cities) Cities() []docker.City {
	return append(c.cities, c.anyCity)
}
