package customjointable

import (
	"time"

	"gorm.io/gorm"

	"ariga.io/atlas-provider-gorm/gormschema"
)

type Person struct {
	ID        int
	Name      string
	Addresses []Address `gorm:"many2many:person_addresses;"`
}

type Address struct {
	ID   int
	Name string
}

type PersonAddress struct {
	PersonID  int `gorm:"primaryKey"`
	AddressID int `gorm:"primaryKey"`
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt
}

type TopCrowdedAddresses struct{}

func (TopCrowdedAddresses) ViewDef() []gormschema.ViewOption {
	return []gormschema.ViewOption{
		gormschema.CreateStmt("CREATE VIEW top_crowded_addresses AS SELECT address_id, COUNT(person_id) AS count FROM person_addresses GROUP BY address_id ORDER BY count DESC LIMIT 10"),
	}
}
