package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name string
	Pets []Pet
}

type TopPetOwner struct{}

func (TopPetOwner) ViewDef(db *gorm.DB) gorm.ViewOption {
	return gorm.ViewOption{
		Query: db.
			Table("users").
			Select("users.id, count(pets.id) as pet_count").
			Joins("left join pets on pets.user_id = users.id").
			Group("users.id").
			Order("pet_count desc").
			Limit(5),
	}
}
