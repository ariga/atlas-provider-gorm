package gormschema

import (
	"strings"
	"testing"

	"ariga.io/atlas-provider-gorm/internal/testdata/models"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestConfig(t *testing.T) {
	l := New("sqlite", WithConfig(
		&gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		},
	))
	sql, err := l.Load(models.Pet{}, models.User{})
	require.NoError(t, err)
	require.Contains(t, sql, "CREATE TABLE `pets`")
	require.Contains(t, sql, "CREATE TABLE `users`")
	require.NotContains(t, sql, "FOREIGN KEY")
}

func TestPostgresConfig(t *testing.T) {
	l := New("postgres", WithConfig(
		&gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: false, // making sure circular FK constraint migration is tested
		},
	))
	sql, err := l.Load(models.Pet{}, models.User{}, models.Toy{})
	require.NoError(t, err)
	require.Contains(t, sql, "CREATE TABLE \"pets\"")
	require.Contains(t, sql, "CREATE TABLE \"users\"")
	require.Contains(t, sql, "CREATE TABLE \"toys\"")

	// Verify FK constraint is moved to the end
	stmts := strings.Split(strings.TrimSpace(sql), "\n")
	require.Contains(t, stmts[len(stmts)-1], "ALTER TABLE \"pets\" ADD CONSTRAINT")
	require.Contains(t, stmts[len(stmts)-1], "FOREIGN KEY")
}
