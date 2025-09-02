package models

import (
	"gorm.io/gorm"

	"ariga.io/atlas-provider-gorm/gormschema"
)

type User struct {
	gorm.Model
	Name    string
	Age     int
	Pets    []Pet
	Hobbies []Hobby `gorm:"many2many:user_hobbies;"`
}

type Hobby struct {
	ID    uint
	Name  string
	Users []User `gorm:"many2many:user_hobbies;"`
}

type WorkingAgedUsers struct {
	Name string
	Age  int
}

func (WorkingAgedUsers) ViewDef(dialect string) []gormschema.ViewOption {
	if dialect == "spanner" {
		return []gormschema.ViewOption{
			gormschema.CreateStmt("CREATE OR REPLACE VIEW working_aged_users\nSQL SECURITY INVOKER\nAS\nSELECT u.name, u.age\nFROM users AS u\nWHERE u.age BETWEEN 18 AND 65"),
		}
	}
	return []gormschema.ViewOption{
		gormschema.BuildStmt(func(db *gorm.DB) *gorm.DB {
			return db.Model(&User{}).Where("age BETWEEN 18 AND 65").Select("name, age")
		}),
	}
}
