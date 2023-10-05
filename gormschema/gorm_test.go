package gormschema

import (
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
