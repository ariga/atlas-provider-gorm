package gormschema

import (
	"testing"

	"ariga.io/atlas-go-sdk/recordriver"
	"ariga.io/atlas-provider-gorm/internal/testdata/models"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestSQLiteConfig(t *testing.T) {
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
	resetSession(t)
	l = New("sqlite", WithConfig(&gorm.Config{}))
	sql, err = l.Load(models.Pet{}, models.User{})
	// Circular foreign keys are not supported in sqlite
	require.Errorf(t, err, "invalid DDL")
}

func TestPostgreSQLConfig(t *testing.T) {
	l := New("postgres", WithConfig(&gorm.Config{}))
	sql, err := l.Load(models.Location{}, models.Event{})
	require.NoError(t, err)
	require.Contains(t, sql, `CREATE TABLE "events"`)
	require.Contains(t, sql, `CREATE UNIQUE INDEX IF NOT EXISTS "idx_events_location_id"`)
	require.Contains(t, sql, `CREATE TABLE "locations"`)
	require.Contains(t, sql, `CREATE UNIQUE INDEX IF NOT EXISTS "idx_locations_event_id"`)
	require.Contains(t, sql, `ALTER TABLE "events" ADD CONSTRAINT "fk_locations_event" FOREIGN KEY ("locationId")`)
	require.Contains(t, sql, `ALTER TABLE "locations" ADD CONSTRAINT "fk_events_location"`)
	resetSession(t)
	l = New("postgres", WithConfig(
		&gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		},
	))
	sql, err = l.Load(models.Location{}, models.Event{})
	require.NoError(t, err)
	require.Contains(t, sql, `CREATE TABLE "events"`)
	require.Contains(t, sql, `CREATE TABLE "locations"`)
	require.Contains(t, sql, `CREATE UNIQUE INDEX IF NOT EXISTS "idx_locations_event_id"`)
	require.NotContains(t, sql, "FOREIGN KEY")
}

func TestMySQLConfig(t *testing.T) {
	l := New("mysql", WithConfig(&gorm.Config{}))
	sql, err := l.Load(models.Location{}, models.Event{})
	require.NoError(t, err)
	require.Contains(t, sql, "CREATE TABLE `events`")
	require.Contains(t, sql, "CREATE TABLE `locations`")
	require.Contains(t, sql, "ALTER TABLE `events` ADD CONSTRAINT `fk_locations_event`")
	require.Contains(t, sql, "ALTER TABLE `locations` ADD CONSTRAINT `fk_events_location`")
	resetSession(t)
	l = New("mysql", WithConfig(
		&gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		},
	))
	sql, err = l.Load(models.Location{}, models.Event{})
	require.NoError(t, err)
	require.Contains(t, sql, "CREATE TABLE `events`")
	require.Contains(t, sql, "CREATE TABLE `locations`")
	require.NotContains(t, sql, "FOREIGN KEY")
}

func resetSession(t *testing.T) {
	sess, ok := recordriver.Session("gorm")
	require.True(t, ok)
	sess.Statements = []string{}
}
