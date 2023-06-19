package gormschema

import (
	"database/sql"
	"database/sql/driver"
	"fmt"

	"ariga.io/atlas-provider-gorm/internal/recordriver"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func New(dialect string) *loader {
	return &loader{dialect: dialect}
}

// loader is a loader for gorm schema.
type loader struct {
	dialect string
}

func (l *loader) Load(models ...any) (string, error) {
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
	db, err := gorm.Open(di, &gorm.Config{})
	if err != nil {
		return "", err
	}
	if err := db.Migrator().CreateTable(models...); err != nil {
		return "", err
	}
	his, ok := recordriver.Session("gorm")
	if !ok {
		return "", err
	}
	return his.Stmts(), nil
}