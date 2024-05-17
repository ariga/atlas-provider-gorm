package models

import (
	"gorm.io/gorm"

	"ariga.io/atlas-provider-gorm/gormschema"
)

type Pet struct {
	gorm.Model
	Name   string
	User   User
	UserID uint
}

type TopPetOwner struct {
	Name     string
	PetCount int
}

func (TopPetOwner) ViewDef(dialect string) []gormschema.ViewOption {
	var stmt string
	switch dialect {
	case "mysql":
		stmt = "CREATE VIEW top_pet_owners AS SELECT user_id, COUNT(id) AS pet_count FROM pets GROUP BY user_id ORDER BY pet_count DESC LIMIT 10"
	case "postgres":
		stmt = "CREATE VIEW top_pet_owners AS SELECT user_id, COUNT(id) AS pet_count FROM pets GROUP BY user_id ORDER BY pet_count DESC LIMIT 10"
	case "sqlite":
		stmt = "CREATE VIEW top_pet_owners AS SELECT user_id, COUNT(id) AS pet_count FROM pets GROUP BY user_id ORDER BY pet_count DESC LIMIT 10"
	case "sqlserver":
		stmt = "CREATE VIEW top_pet_owners AS SELECT user_id, COUNT(id) AS pet_count FROM pets GROUP BY user_id ORDER BY pet_count DESC OFFSET 0 ROWS FETCH NEXT 10 ROWS ONLY"
	}
	return []gormschema.ViewOption{gormschema.CreateStmt(stmt)}
}
