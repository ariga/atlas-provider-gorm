package models

import (
	"gorm.io/gorm"

	"ariga.io/atlas-provider-gorm/gormschema"
)

type User struct {
	gorm.Model
	Name string
	Age  int
	Pets []Pet
}

type WorkingAgedUsers struct {
	Name string
	Age  int
}

func (WorkingAgedUsers) ViewDef(dialect string) []gormschema.ViewOption {
	return []gormschema.ViewOption{
		gormschema.BuildStmt(func(db *gorm.DB) *gorm.DB {
			return db.Model(&User{}).Where("age BETWEEN 18 AND 65").Select("name, age")
		}),
	}
}
