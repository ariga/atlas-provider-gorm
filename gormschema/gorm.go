package gormschema

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"slices"

	"ariga.io/atlas-go-sdk/recordriver"
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
	var (
		views  []ViewDefiner
		tables []any
	)
	for _, obj := range models {
		switch view := obj.(type) {
		case ViewDefiner:
			views = append(views, view)
		default:
			tables = append(tables, obj)
		}
	}
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
	if err = db.AutoMigrate(tables...); err != nil {
		return "", err
	}
	db, err = gorm.Open(dialector{Dialector: di}, l.config)
	if err != nil {
		return "", err
	}
	cm, ok := db.Migrator().(*migrator)
	if !ok {
		return "", fmt.Errorf("unexpected migrator type: %T", db.Migrator())
	}
	if err = cm.CreateViews(views); err != nil {
		return "", err
	}
	if !l.config.DisableForeignKeyConstraintWhenMigrating && l.dialect != "sqlite" {
		if err = cm.CreateConstraints(tables); err != nil {
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

// Migrator returns a new gorm.Migrator, which can be used to extend the default migrator,
// helping to create constraints and views ...
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

// CreateViews creates the given "view-based" models
func (m *migrator) CreateViews(views []ViewDefiner) error {
	for _, view := range views {
		viewName := m.DB.Config.NamingStrategy.TableName(indirect(reflect.TypeOf(view)).Name())
		if namer, ok := view.(interface {
			TableName() string
		}); ok {
			viewName = namer.TableName()
		}
		viewBuilder := &viewBuilder{
			db:       m.DB,
			viewName: viewName,
		}
		for _, opt := range view.ViewDef(m.Dialector.Name()) {
			opt(viewBuilder)
		}
		if err := m.DB.Exec(viewBuilder.createStmt).Error; err != nil {
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

func indirect(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

type (
	// ViewOption configures a viewBuilder.
	ViewOption func(*viewBuilder)
	// ViewDefiner defines a view.
	ViewDefiner interface {
		ViewDef(driver string) []ViewOption
	}
	viewBuilder struct {
		db         *gorm.DB
		createStmt string
		// viewName is only used for the BuildStmt option.
		// BuildStmt returns only a subquery; viewName helps to create a full CREATE VIEW statement.
		viewName string
	}
)

// CreateStmt accepts raw SQL with args to create a CREATE VIEW statement.
func CreateStmt(sql string, args ...any) ViewOption {
	return func(b *viewBuilder) {
		b.createStmt = b.db.ToSQL(func(tx *gorm.DB) *gorm.DB {
			return tx.Exec(sql, args...)
		})
	}
}

// BuildStmt accepts a function with gorm query builder to create a CREATE VIEW statement.
// With this option, the view's name will be the same as the model's table name
func BuildStmt(fn func(db *gorm.DB) *gorm.DB) ViewOption {
	return func(b *viewBuilder) {
		vd := b.db.ToSQL(func(tx *gorm.DB) *gorm.DB {
			return fn(tx).
				Unscoped(). // Skip gorm deleted_at filtering.
				Find(nil)   // Execute the query and convert it to SQL.
		})

		b.createStmt = fmt.Sprintf("CREATE VIEW %s AS %s", b.viewName, vd)
	}
}
