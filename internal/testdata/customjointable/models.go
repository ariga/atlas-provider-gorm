package customjointable

import (
	"time"

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

func (TopCrowdedAddresses) ViewDef(db *gorm.DB) gorm.ViewOption {
	return gorm.ViewOption{
		Query: db.
			Table("addresses").
			Select("addresses.id, addresses.name, count(person_addresses.person_id) as person_count").
			Joins("left join person_addresses on person_addresses.address_id = addresses.id").
			Group("addresses.id").
			Order("person_count desc").
			Limit(10),
	}
}
