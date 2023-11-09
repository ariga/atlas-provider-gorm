package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name string
	Pets []Pet
}

type Pet struct {
	gorm.Model
	ID     uint
	Name   string
	User   User
	UserID uint
	Toy    *Toy `gorm:"foreignKey:toy_id;references:id;OnUpdate:CASCADE,OnDelete:CASCADE"`
	ToyID  uint
}

type Toy struct {
	ID    uint
	Name  string
	PetID uint
	Pet   *Pet `gorm:"foreignKey:pet_id;references:id;OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type House struct {
	ID   uint
	Name string
}
