package gormschema

import (
	"database/sql"
	"database/sql/driver"
	"fmt"

	"ariga.io/atlas-go-sdk/recordriver"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormMig "gorm.io/gorm/migrator"
)

// New returns a new Loader.
func New(dialect string, opts ...Option) *Loader {
	l := &Loader{dialect: dialect, config: &gorm.Config{}}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

type (
	// Loader is a Loader for gorm schema.
	Loader struct {
		dialect                           string
		createConstraintsAfterCreateTable bool
		config                            *gorm.Config
	}
	// Option configures the Loader.
	Option func(*Loader)
)

// WithConfig sets the gorm config.
func WithConfig(cfg *gorm.Config) Option {
	return func(l *Loader) {
		l.config = cfg
	}
}

// WithCreateConstraintsAfterCreateTable sets the createConstraintsAfterCreateTable config.
func WithCreateConstraintsAfterCreateTable(b bool) Option {
	return func(l *Loader) {
		l.createConstraintsAfterCreateTable = b
	}
}

// Load loads the models and returns the DDL statements representing the schema.
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
	db, err := gorm.Open(di, l.config)
	if err != nil {
		return "", err
	}
	if l.createConstraintsAfterCreateTable {
		db.Config.DisableForeignKeyConstraintWhenMigrating = true
	}
	if err = db.AutoMigrate(models...); err != nil {
		return "", err
	}
	if l.createConstraintsAfterCreateTable {
		db, err = gorm.Open(dialector{
			Dialector: di,
		}, l.config)
		if err != nil {
			return "", err
		}
		cm, ok := db.Migrator().(*migrator)
		if !ok {
			return "", err
		}
		if err = cm.CreateConstraints(models); err != nil {
			return "", err
		}
	}
	s, ok := recordriver.Session("gorm")
	if !ok {
		return "", fmt.Errorf("gorm db session not found")
	}
	return s.Stmts(), nil
}

type migrator struct {
	gormMig.Migrator
	dialectMigrator gorm.Migrator
}

type dialector struct {
	gorm.Dialector
}

// Migrator returns a new gorm.Migrator which can be used to automatically create all Constraints
// on existing tables.
func (d dialector) Migrator(db *gorm.DB) gorm.Migrator {
	return &migrator{
		Migrator: gormMig.Migrator{
			Config: gormMig.Config{
				DB:        db,
				Dialector: d,
			},
		},
		dialectMigrator: d.Dialector.Migrator(db),
	}
}

// HasTable always returns `true`. By returning `true`, gorm.Migrator will try to alter the table to add constraints.
func (m *migrator) HasTable(dst interface{}) bool {
	return true
}

// CreateConstraints detects constraints on the given model and creates them using `m.dialectMigrator`.
func (m *migrator) CreateConstraints(models []interface{}) error {
	for _, model := range m.ReorderModels(models, true) {
		err := m.Migrator.RunWithValue(model, func(stmt *gorm.Statement) error {
			for _, rel := range stmt.Schema.Relationships.Relations {
				if rel.Field.IgnoreMigration {
					continue
				}
				if constraint := rel.ParseConstraint(); constraint != nil &&
					constraint.Schema == stmt.Schema {
					if err := m.dialectMigrator.CreateConstraint(model, constraint.Name); err != nil {
						return err
					}
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}
