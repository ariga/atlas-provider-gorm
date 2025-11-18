package spanner

import (
	"sync"

	"ariga.io/atlas-provider-gorm/gormschema"
	spannergorm "github.com/googleapis/go-gorm-spanner"
	"gorm.io/gorm"
)

var registerOnce sync.Once

func init() {
	Register()
}

// Register makes the Spanner dialector available to gormschema.Loaders.
// Users can import this package for its side effects:
//
//	import _ "ariga.io/atlas-provider-gorm/gormschema/dialect/spanner"
func Register() {
	registerOnce.Do(func() {
		gormschema.RegisterDialector("spanner", func(*gormschema.Loader) (gorm.Dialector, error) {
			return spannergorm.New(spannergorm.Config{
				DriverName:                 "recordriver",
				DisableAutoMigrateBatching: true,
				DSN:                        "gorm",
			}), nil
		})
	})
}
