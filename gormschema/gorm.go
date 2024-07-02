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

type (
	// Loader is a Loader for gorm schema.
	Loader struct {
		dialect           string
		config            *gorm.Config
		beforeAutoMigrate []func(*gorm.DB) error
	}
	// Option configures the Loader.
	Option func(*Loader)
	// ViewOption implemented by VIEW's related options
	ViewOption interface {
		isViewOption()
		apply(*schemaBuilder)
	}
	// TriggerOption implemented by TRIGGER's related options
	TriggerOption interface {
		isTriggerOption()
		apply(*schemaBuilder)
	}
	// Trigger defines a trigger.
	Trigger struct {
		opts []TriggerOption
	}
	// ViewDefiner defines a view.
	ViewDefiner interface {
		ViewDef(dialect string) []ViewOption
	}
	// schemaOption configures the schemaBuilder.
	schemaOption  func(*schemaBuilder)
	schemaBuilder struct {
		db         *gorm.DB
		createStmt string
		// viewName is only used for the BuildStmt option.
		// BuildStmt returns only a subquery; viewName helps to create a full CREATE VIEW statement.
		viewName string
	}
)

// WithConfig sets the gorm config.
func WithConfig(cfg *gorm.Config) Option {
	return func(l *Loader) {
		l.config = cfg
	}
}

// WithJoinTable sets up a join table for the given model and field.
// Deprecated: put the join tables alongside the models in the Load call.
func WithJoinTable(model any, field string, jointable any) Option {
	return func(l *Loader) {
		l.beforeAutoMigrate = append(l.beforeAutoMigrate, func(db *gorm.DB) error {
			return db.SetupJoinTable(model, field, jointable)
		})
	}
}

// New returns a new Loader.
func New(dialect string, opts ...Option) *Loader {
	l := &Loader{dialect: dialect, config: &gorm.Config{}}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

// NewTrigger receives a list of TriggerOption to build a Trigger.
func NewTrigger(opts ...TriggerOption) Trigger {
	return Trigger{opts: opts}
}

func (s schemaOption) apply(b *schemaBuilder) {
	s(b)
}

func (schemaOption) isViewOption()    {}
func (schemaOption) isTriggerOption() {}

// CreateStmt accepts raw SQL to create a view or trigger
func CreateStmt(stmt string) interface {
	ViewOption
	TriggerOption
} {
	return schemaOption(func(b *schemaBuilder) {
		b.createStmt = stmt
	})
}

// BuildStmt accepts a function with gorm query builder to create a CREATE VIEW statement.
// With this option, the view's name will be the same as the model's table name
func BuildStmt(fn func(db *gorm.DB) *gorm.DB) ViewOption {
	return schemaOption(func(b *schemaBuilder) {
		vd := b.db.ToSQL(func(tx *gorm.DB) *gorm.DB {
			return fn(tx).
				Unscoped(). // Skip gorm deleted_at filtering.
				Find(nil)   // Execute the query and convert it to SQL.
		})
		b.createStmt = fmt.Sprintf("CREATE VIEW %s AS %s", b.viewName, vd)
	})
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
	cdb, err := gorm.Open(dialector{Dialector: di}, l.config)
	if err != nil {
		return "", err
	}
	cm, ok := cdb.Migrator().(*migrator)
	if !ok {
		return "", fmt.Errorf("unexpected migrator type: %T", db.Migrator())
	}
	if err = cm.setupJoinTables(tables...); err != nil {
		return "", err
	}
	orderedTables, err := cm.orderModels(tables...)
	if err != nil {
		return "", err
	}
	if err = db.AutoMigrate(orderedTables...); err != nil {
		return "", err
	}
	if err = cm.CreateViews(views); err != nil {
		return "", err
	}
	if err = cm.CreateTriggers(models); err != nil {
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

// setupJoinTables helps to determine custom join tables present in the model list and sets them up.
func (m *migrator) setupJoinTables(models ...any) error {
	var dbNameModelMap = make(map[string]any)
	for _, model := range models {
		err := m.RunWithValue(model, func(stmt *gorm.Statement) error {
			dbNameModelMap[stmt.Schema.Table] = model
			return nil
		})
		if err != nil {
			return err
		}
	}
	for _, model := range m.ReorderModels(models, false) {
		err := m.RunWithValue(model, func(stmt *gorm.Statement) error {
			for _, rel := range stmt.Schema.Relationships.Relations {
				if rel.Field.IgnoreMigration || rel.JoinTable == nil {
					continue
				}
				if joinTable, ok := dbNameModelMap[rel.JoinTable.Name]; ok {
					if err := m.DB.SetupJoinTable(model, rel.Field.Name, joinTable); err != nil {
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
		schemaBuilder := &schemaBuilder{
			db:       m.DB,
			viewName: viewName,
		}
		for _, opt := range view.ViewDef(m.Dialector.Name()) {
			opt.apply(schemaBuilder)
		}
		if err := m.DB.Exec(schemaBuilder.createStmt).Error; err != nil {
			return err
		}
	}
	return nil
}

// orderModels places join tables at the end of the list of models (if any),
// which helps GORM resolve m2m relationships correctly.
func (m *migrator) orderModels(models ...any) ([]any, error) {
	var (
		joinTableDBNames = make(map[string]bool)
		otherTables      []any
		joinTables       []any
	)
	for _, model := range models {
		err := m.RunWithValue(model, func(stmt *gorm.Statement) error {
			for _, rel := range stmt.Schema.Relationships.Relations {
				if rel.JoinTable != nil {
					joinTableDBNames[rel.JoinTable.Table] = true
					return nil
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	for _, model := range models {
		err := m.RunWithValue(model, func(stmt *gorm.Statement) error {
			if joinTableDBNames[stmt.Schema.Table] {
				joinTables = append(joinTables, model)
			} else {
				otherTables = append(otherTables, model)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return append(otherTables, joinTables...), nil
}

// CreateTriggers creates the triggers for the given models.
func (m *migrator) CreateTriggers(models []any) error {
	for _, model := range models {
		if md, ok := model.(interface {
			Triggers(string) []Trigger
		}); ok {
			for _, trigger := range md.Triggers(m.Dialector.Name()) {
				schemaBuilder := &schemaBuilder{
					db: m.DB,
				}
				for _, opt := range trigger.opts {
					opt.apply(schemaBuilder)
					if err := m.DB.Exec(schemaBuilder.createStmt).Error; err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func indirect(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}
