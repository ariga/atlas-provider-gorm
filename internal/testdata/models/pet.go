package models

import (
	"ariga.io/atlas-provider-gorm/gormschema"
	"gorm.io/gorm"
)

type Pet struct {
	gorm.Model
	Name   string
	User   User
	UserID uint
}

type BotlTracker struct {
	ID   uint
	Name string
}

func (BotlTracker) TableName() string {
	return "botl_tracker_custom_name"
}

func (BotlTracker) ViewDef() []gormschema.ViewOption {
	return []gormschema.ViewOption{
		gormschema.CreateStmt("CREATE VIEW botl_tracker_custom_name AS SELECT id, name FROM pets WHERE name LIKE ?", "botl%"),
	}
}
