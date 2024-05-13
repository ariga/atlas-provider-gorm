package models

import (
	"ariga.io/atlas-provider-gorm/gormschema"
	"gorm.io/gorm"
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

func (WorkingAgedUsers) ViewDef() gormschema.ViewDef {
	return gormschema.ViewDef{
		Def: "SELECT name, age FROM users WHERE age BETWEEN 18 AND 65",
	}
}
