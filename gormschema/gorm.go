package gormschema

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"gorm.io/gorm/schema"

	"ariga.io/atlas-go-sdk/recordriver"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// New returns a new Loader.
func New(dialect string, schema string) *Loader {
	return &Loader{dialect: dialect, schema: schema}
}

// Loader is a Loader for gorm schema.
type Loader struct {
	dialect string
	schema  string
}

func (l *Loader) Load(models ...any) (string, error) {
	var di gorm.Dialector
	switch l.dialect {
	case "sqlite":
		rd, err := sql.Open("recordriver", "gorm")
		if err != nil {
			return "", err
		}
		di = sqlite.Dialector{Conn: rd}
		recordriver.SetResponse("gorm", "select sqlite_version()", &recordriver.Response{
			Cols: []string{"sqlite_version()"},
			Data: [][]driver.Value{{"3.30.1"}},
		})
	case "mysql":
		di = mysql.New(mysql.Config{
			DriverName: "recordriver",
			DSN:        "gorm",
		})
		recordriver.SetResponse("gorm", "SELECT VERSION()", &recordriver.Response{
			Cols: []string{"VERSION()"},
			Data: [][]driver.Value{{"8.0.24"}},
		})
	case "postgres":
		di = postgres.New(postgres.Config{
			DriverName: "recordriver",
			DSN:        "gorm",
		})
	default:
		return "", fmt.Errorf("unsupported engine: %s", l.dialect)
	}
	var config gorm.Config
	if l.schema != "" {
		config.NamingStrategy = schema.NamingStrategy{
			TablePrefix: l.schema + ".",
		}
	}
	db, err := gorm.Open(di, &config)
	if err != nil {
		return "", err
	}
	if err := db.AutoMigrate(models...); err != nil {
		return "", err
	}
	s, ok := recordriver.Session("gorm")
	if !ok {
		return "", err
	}
	return s.Stmts(), nil
}
