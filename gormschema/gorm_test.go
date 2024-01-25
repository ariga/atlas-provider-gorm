package gormschema

import (
	"os"
	"testing"

	"ariga.io/atlas-go-sdk/recordriver"
	ckmodels "ariga.io/atlas-provider-gorm/internal/testdata/circularfks"
	"ariga.io/atlas-provider-gorm/internal/testdata/models"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestSQLiteConfig(t *testing.T) {
	resetSession()
	l := New("sqlite")
	sql, err := l.Load(models.Pet{}, models.User{}, ckmodels.Event{}, ckmodels.Location{})
	require.NoError(t, err)
	requireEqualContent(t, sql, "testdata/sqlite_default")
	resetSession()
	l = New("sqlite", WithConfig(&gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	}))
	sql, err = l.Load(models.Pet{}, models.User{})
	require.NoError(t, err)
	requireEqualContent(t, sql, "testdata/sqlite_no_fk")
	resetSession()
}

func TestPostgreSQLConfig(t *testing.T) {
	resetSession()
	l := New("postgres")
	sql, err := l.Load(ckmodels.Location{}, ckmodels.Event{}, models.User{}, models.Pet{})
	require.NoError(t, err)
	requireEqualContent(t, sql, "testdata/postgresql_default")
	resetSession()
	l = New("postgres", WithConfig(
		&gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		}))
	sql, err = l.Load(ckmodels.Location{}, ckmodels.Event{})
	require.NoError(t, err)
	requireEqualContent(t, sql, "testdata/postgresql_no_fk")
}

func TestMySQLConfig(t *testing.T) {
	resetSession()
	l := New("mysql")
	sql, err := l.Load(ckmodels.Location{}, ckmodels.Event{}, models.User{}, models.Pet{})
	require.NoError(t, err)
	requireEqualContent(t, sql, "testdata/mysql_default")
	resetSession()
	l = New("mysql", WithConfig(
		&gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		},
	))
	sql, err = l.Load(ckmodels.Location{}, ckmodels.Event{})
	require.NoError(t, err)
	requireEqualContent(t, sql, "testdata/mysql_no_fk")
}

func TestSQLServerConfig(t *testing.T) {
	resetSession()
	l := New("sqlserver")
	sql, err := l.Load(ckmodels.Location{}, ckmodels.Event{}, models.User{}, models.Pet{})
	require.NoError(t, err)
	requireEqualContent(t, sql, "testdata/sqlserver_default")
	resetSession()
	l = New("sqlserver", WithConfig(
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

func requireEqualContent(t *testing.T, expected, fileName string) {
	buf, err := os.ReadFile(fileName)
	require.NoError(t, err)
	require.Equal(t, expected, string(buf))
}
