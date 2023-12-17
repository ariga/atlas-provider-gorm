package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name string
	Pets []Pet
}

type Pet struct {
	gorm.Model
	Name   string
	User   User
	UserID uint
}

type Toy struct {
	ID   uint
	Name string
}

type Location struct {
	LocationID string `gorm:"primaryKey;column:locationId;"`
	EventID    string `gorm:"uniqueIndex;column:eventId;"`
	Event      *Event `gorm:"foreignKey:locationId;references:locationId;OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type Event struct {
	EventID    string    `gorm:"primaryKey;column:eventId;"`
	LocationID string    `gorm:"uniqueIndex;column:locationId;"`
	Location   *Location `gorm:"foreignKey:eventId;references:eventId;OnUpdate:CASCADE,OnDelete:CASCADE"`
}
