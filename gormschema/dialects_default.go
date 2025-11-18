package gormschema

import (
	"database/sql"
	"database/sql/driver"

	"ariga.io/atlas/sdk/recordriver"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

func init() {
	RegisterDialector("sqlite", func(*Loader) (gorm.Dialector, error) {
		rd, err := sql.Open("recordriver", "gorm")
		if err != nil {
			return nil, err
		}
		recordriver.SetResponse("gorm", "select sqlite_version()", &recordriver.Response{
			Cols: []string{"sqlite_version()"},
			Data: [][]driver.Value{{"3.30.1"}},
		})
		return sqlite.Dialector{Conn: rd}, nil
	})
	RegisterDialector("mysql", func(*Loader) (gorm.Dialector, error) {
		recordriver.SetResponse("gorm", "SELECT VERSION()", &recordriver.Response{
			Cols: []string{"VERSION()"},
			Data: [][]driver.Value{{"8.0.24"}},
		})
		return mysql.New(mysql.Config{
			DriverName: "recordriver",
			DSN:        "gorm",
		}), nil
	})
	RegisterDialector("postgres", func(*Loader) (gorm.Dialector, error) {
		return postgres.New(postgres.Config{
			DriverName: "recordriver",
			DSN:        "gorm",
		}), nil
	})
	RegisterDialector("sqlserver", func(*Loader) (gorm.Dialector, error) {
		return sqlserver.New(sqlserver.Config{
			DriverName: "recordriver",
			DSN:        "gorm",
		}), nil
	})
}
