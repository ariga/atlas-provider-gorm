package gormschema

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"slices"

	"ariga.io/atlas-go-sdk/recordriver"
	"github.com/go-openapi/inflect"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	gormig "gorm.io/gorm/migrator"
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
		dialect           string
		config            *gorm.Config
		beforeAutoMigrate []func(*gorm.DB) error
		afterAutoMigrate  []func(*gorm.DB) error
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
	case "sqlserver":
		di = sqlserver.New(sqlserver.Config{
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
	if l.dialect != "sqlite" {
		db.Config.DisableForeignKeyConstraintWhenMigrating = true
	}
	for _, cb := range l.beforeAutoMigrate {
		if err = cb(db); err != nil {
			return "", err
		}
	}
	if err = db.AutoMigrate(models...); err != nil {
		return "", err
	}
	for _, cb := range l.afterAutoMigrate {
		if err = cb(db); err != nil {
			return "", err
		}
	}
	if !l.config.DisableForeignKeyConstraintWhenMigrating && l.dialect != "sqlite" {
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
		return "", errors.New("gorm db session not found")
	}
	return s.Stmts(), nil
}

type migrator struct {
	gormig.Migrator
	dialectMigrator gorm.Migrator
}

type dialector struct {
	gorm.Dialector
}

// Migrator returns a new gorm.Migrator which can be used to automatically create all Constraints
// on existing tables.
func (d dialector) Migrator(db *gorm.DB) gorm.Migrator {
	return &migrator{
		Migrator: gormig.Migrator{
			Config: gormig.Config{
				DB:        db,
				Dialector: d,
			},
		},
		dialectMigrator: d.Dialector.Migrator(db),
	}
}

// HasTable always returns `true`. By returning `true`, gorm.Migrator will try to alter the table to add constraints.
func (m *migrator) HasTable(dst any) bool {
	return true
}

// CreateConstraints detects constraints on the given model and creates them using `m.dialectMigrator`.
func (m *migrator) CreateConstraints(models []any) error {
	for _, model := range m.ReorderModels(models, true) {
		err := m.Migrator.RunWithValue(model, func(stmt *gorm.Statement) error {

			relationNames := make([]string, 0, len(stmt.Schema.Relationships.Relations))
			for name := range stmt.Schema.Relationships.Relations {
				relationNames = append(relationNames, name)
			}
			// since Relations is a map, the order of the keys is not guaranteed
			// so we sort the keys to make the sql output deterministic
			slices.Sort(relationNames)

			for _, name := range relationNames {
				rel := stmt.Schema.Relationships.Relations[name]

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

// WithJoinTable sets up a join table for the given model and field.
func WithJoinTable(model any, field string, jointable any) Option {
	return func(l *Loader) {
		l.beforeAutoMigrate = append(l.beforeAutoMigrate, func(db *gorm.DB) error {
			return db.SetupJoinTable(model, field, jointable)
		})
	}
}

type (
	view interface {
		ViewDef(*gorm.DB) gorm.ViewOption
	}
)

// WithViews sets up callbacks to create views for the given "view-based" models.
func WithViews(models ...any) Option {
	return func(l *Loader) {
		for _, model := range models {
			if view, ok := model.(view); ok {
				l.afterAutoMigrate = append(l.afterAutoMigrate, func(db *gorm.DB) error {
					viewDef := view.ViewDef(db)
					return db.Migrator().CreateView(inflect.Underscore(indirect(reflect.TypeOf(view)).Name()), gorm.ViewOption{
						Replace:     viewDef.Replace,
						CheckOption: viewDef.CheckOption,
						Query:       viewDef.Query,
					})
				})
			}
		}
	}
}

func indirect(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}
