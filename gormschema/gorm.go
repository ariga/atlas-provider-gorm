package gormschema

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"regexp"
	"strings"

	"ariga.io/atlas-go-sdk/recordriver"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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
		dialect string
		config  *gorm.Config
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
	default:
		return "", fmt.Errorf("unsupported engine: %s", l.dialect)
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
		return "", err
	}

	switch l.dialect {
	case "sqlite":
		// To be implemented
	case "mysql":
		// To be implemented
	case "postgres":
		s.Statements = moveConstraintStatementsLast(s.Statements)
	default:
		return "", fmt.Errorf("unsupported engine: %s", l.dialect)
	}

	return s.Stmts(), nil
}

// Moves FK contraint statements nested within a CREATE TABLE statement out to the end
// Works for PostgreSQL only
func moveConstraintStatementsLast(stmts []string) []string {
	newStmts := []string{}
	contraintStmts := []string{}

	// Regex to match CREATE TABLE X before the first bracket (
	regexPrefix := *regexp.MustCompile(`^[^(]*`)
	// Regex to match content within the brackets
	regexRows := *regexp.MustCompile(`\((.*)\)`)

	for _, stmt := range stmts {
		if !strings.HasPrefix(stmt, "CREATE TABLE") {
			newStmts = append(newStmts, stmt)
			continue
		}


		matchedPrefix := regexPrefix.FindAllStringSubmatch(stmt, -1)
		if len(matchedPrefix) <= 0 && len(matchedPrefix[0]) <= 0 {
			newStmts = append(newStmts, stmt)
			continue
		}
		prefix := matchedPrefix[0][0]
		splitPrefix := strings.Split(strings.TrimSpace(prefix), " ")
		tableName := splitPrefix[len(splitPrefix)-1]

		matchedRows := regexRows.FindAllStringSubmatch(stmt, -1)
		if len(matchedRows) <= 0 && len(matchedRows[0]) <= 1 {
			newStmts = append(newStmts, stmt)
			continue
		}

		rows := strings.Split(matchedRows[0][1], ",")
		newStmt := []string{}
		for _, row := range rows {
			if !strings.Contains(row, "FOREIGN KEY") {
				newStmt = append(newStmt, row)
				continue
			}

			contraintStmts = append(
				contraintStmts,
				fmt.Sprintf("ALTER TABLE %v ADD %v", tableName, row),
			)
		}

		newStmtString := prefix + "(" + strings.Join(newStmt[:], ",") + ")"
		newStmts = append(newStmts, newStmtString)
	}

	return append(newStmts, contraintStmts...)
}
