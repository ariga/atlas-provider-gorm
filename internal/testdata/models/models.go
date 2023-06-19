package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name string
}

type Pet struct {
	ID   uint `gorm:"column:foo"`
	Name string
}

type Toy struct {
	ID   uint
	Name string
}
