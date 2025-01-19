package database

import (
	"gorm.io/gorm"
)

func StartAndEndWith(start, end string, g *gorm.DB) *gorm.DB {
	return g.Where("path ~ ?", start+".*."+end)
}

func StartWith(start string, g *gorm.DB) *gorm.DB {
	return g.Where("path ~ ?", start+".*")
}

func EndWith(end string, g *gorm.DB) *gorm.DB {
	return g.Where("path ~ ?", ".*."+end)
}

func Contains(contains string, g *gorm.DB) *gorm.DB {
	return g.Where("path ~ ?", ".*"+contains+".*")
}

func Equals(equals string, g *gorm.DB) *gorm.DB {
	return g.Where("path = ?", equals)
}

func NotEquals(notEquals string, g *gorm.DB) *gorm.DB {
	return g.Where("path != ?", notEquals)
}
