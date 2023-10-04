package gormschema

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"

	"ariga.io/atlas-go-sdk/recordriver"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// New returns a new Loader.
func New(dialect string, config ...func(*gorm.Config)) *Loader {
	gormConfig := &gorm.Config{}

	for _, c := range config {
		c(gormConfig)
	}

	return &Loader{
		dialect: dialect,
		config:  gormConfig,
	}
}

// Loader is a Loader for gorm schema.
type Loader struct {
	dialect string
	config  *gorm.Config
}

func (l *Loader) Load(models ...interface{}) (string, error) {
	di, err := l.getDialector()
	if err != nil {
		return "", err
	}

	db, err := gorm.Open(di, l.config)
	if err != nil {
		return "", err
	}

	if err := db.AutoMigrate(models...); err != nil {
		return "", err
	}

	s, ok := recordriver.Session("gorm")
	if !ok {
		return "", errors.New("failed to retrieve recordriver session")
	}
	return s.Stmts(), nil
}

func WithForeignKeys(b bool) func(*gorm.Config) {
	return func(c *gorm.Config) {
		c.DisableForeignKeyConstraintWhenMigrating = !b
	}
}

func (l *Loader) getDialector() (gorm.Dialector, error) {
	switch l.dialect {
	case "sqlite":
		rd, err := sql.Open("recordriver", "gorm")
		if err != nil {
			return nil, err
		}
		recordriver.SetResponse("gorm", "select sqlite_version()", &recordriver.Response{
			Cols: []string{"sqlite_version()"},
			Data: [][]driver.Value{{"3.30.1"}},
		})
		return sqlite.Dialector{Conn: rd}, nil
	case "mysql":
		recordriver.SetResponse("gorm", "SELECT VERSION()", &recordriver.Response{
			Cols: []string{"VERSION()"},
			Data: [][]driver.Value{{"8.0.24"}},
		})
		return mysql.New(mysql.Config{
			DriverName: "recordriver",
			DSN:        "gorm",
		}), nil
	case "postgres":
		return postgres.New(postgres.Config{
			DriverName: "recordriver",
			DSN:        "gorm",
		}), nil
	default:
		return nil, fmt.Errorf("unsupported engine: %s", l.dialect)
	}
}
