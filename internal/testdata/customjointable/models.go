package customjointable

import (
	"time"

	"ariga.io/atlas-provider-gorm/gormschema"
	"gorm.io/gorm"
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

func (TopCrowdedAddresses) ViewDef() gormschema.ViewDef {
	return gormschema.ViewDef{
		Def: "SELECT address_id, COUNT(person_id) AS count FROM person_addresses GROUP BY address_id ORDER BY count DESC LIMIT 10",
	}
}
