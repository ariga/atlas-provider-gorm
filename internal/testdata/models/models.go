package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name string
}

type Pet struct {
	ID   uint
	Name string `gorm:"column:pet_name"`
}

type Toy struct {
	ID   uint
	Name string
}
