package gormschema_test

import (
	"os"
	"testing"

	"ariga.io/atlas-go-sdk/recordriver"
	"ariga.io/atlas-provider-gorm/gormschema"
	ckmodels "ariga.io/atlas-provider-gorm/internal/testdata/circularfks"
	"ariga.io/atlas-provider-gorm/internal/testdata/customjointable"
	"ariga.io/atlas-provider-gorm/internal/testdata/models"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestSQLiteConfig(t *testing.T) {
	resetSession()
	l := gormschema.New("sqlite")
	sql, err := l.Load(models.Pet{}, models.User{}, ckmodels.Event{}, ckmodels.Location{})
	require.NoError(t, err)
	requireEqualContent(t, sql, "testdata/sqlite_default")
	resetSession()
	l = gormschema.New("sqlite", gormschema.WithConfig(&gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	}))
	sql, err = l.Load(models.Pet{}, models.User{})
	require.NoError(t, err)
	requireEqualContent(t, sql, "testdata/sqlite_no_fk")
	resetSession()
}

func TestPostgreSQLConfig(t *testing.T) {
	resetSession()
	l := gormschema.New("postgres")
	sql, err := l.Load(ckmodels.Location{}, ckmodels.Event{}, models.User{}, models.Pet{})
	require.NoError(t, err)
	requireEqualContent(t, sql, "testdata/postgresql_default")
	resetSession()
	l = gormschema.New("postgres", gormschema.WithConfig(
		&gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		}))
	sql, err = l.Load(ckmodels.Location{}, ckmodels.Event{})
	require.NoError(t, err)
	requireEqualContent(t, sql, "testdata/postgresql_no_fk")
}

func TestMySQLConfig(t *testing.T) {
	resetSession()
	l := gormschema.New("mysql")
	sql, err := l.Load(ckmodels.Location{}, ckmodels.Event{}, models.User{}, models.Pet{})
	require.NoError(t, err)
	requireEqualContent(t, sql, "testdata/mysql_default")
	resetSession()
	l = gormschema.New("mysql", gormschema.WithConfig(
		&gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		},
	))
	sql, err = l.Load(ckmodels.Location{}, ckmodels.Event{})
	require.NoError(t, err)
	requireEqualContent(t, sql, "testdata/mysql_no_fk")
	resetSession()
	l = gormschema.New("mysql", gormschema.WithJoinTable(&customjointable.Person{}, "Addresses", &customjointable.PersonAddress{}))
	sql, err = l.Load(customjointable.Address{}, customjointable.Person{})
	require.NoError(t, err)
	requireEqualContent(t, sql, "testdata/mysql_custom_join_table")
}

func TestSQLServerConfig(t *testing.T) {
	resetSession()
	l := gormschema.New("sqlserver")
	sql, err := l.Load(ckmodels.Location{}, ckmodels.Event{}, models.User{}, models.Pet{})
	require.NoError(t, err)
	requireEqualContent(t, sql, "testdata/sqlserver_default")
	resetSession()
	l = gormschema.New("sqlserver", gormschema.WithConfig(
		&gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		}))
	sql, err = l.Load(ckmodels.Location{}, ckmodels.Event{})
	require.NoError(t, err)
	requireEqualContent(t, sql, "testdata/sqlserver_no_fk")
}

func resetSession() {
	sess, ok := recordriver.Session("gorm")
	if ok {
		sess.Statements = nil
	}
}

func requireEqualContent(t *testing.T, actual, fileName string) {
	buf, err := os.ReadFile(fileName)
	require.NoError(t, err)
	require.Equal(t, string(buf), actual)
}
