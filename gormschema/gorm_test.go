package gormschema

import (
	"strings"
	"testing"

	"ariga.io/atlas-go-sdk/recordriver"
	ckmodels "ariga.io/atlas-provider-gorm/internal/testdata/circularfks"
	"ariga.io/atlas-provider-gorm/internal/testdata/models"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestSQLiteConfig(t *testing.T) {
	resetSession(t)
	l := New("sqlite")
	sql, err := l.Load(models.Pet{}, models.User{}, ckmodels.Event{}, ckmodels.Location{})
	require.NoError(t, err)
	stmts := strings.Split(sql, "\n")
	require.Contains(t, stmts[0], "CREATE TABLE `users`")
	require.Contains(t, stmts[2], "CREATE TABLE `pets`")
	require.Contains(t, stmts[2], "CONSTRAINT `fk_users_pets` FOREIGN KEY (`user_id`)")
	require.Contains(t, stmts[4], "CREATE TABLE `locations`")
	require.Contains(t, stmts[4], "CONSTRAINT `fk_events_location` FOREIGN KEY (`eventId`)")
	require.Contains(t, stmts[6], "CREATE TABLE `events`")
	require.Contains(t, stmts[6], "CONSTRAINT `fk_locations_event` FOREIGN KEY (`locationId`)")
	resetSession(t)
	l = New("sqlite", WithConfig(&gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	}))
	sql, err = l.Load(models.Pet{}, models.User{})
	require.NoError(t, err)
	stmts = strings.Split(sql, "\n")
	require.Contains(t, stmts[0], "CREATE TABLE `users`")
	require.Contains(t, stmts[2], "CREATE TABLE `pets`")
	require.NotContains(t, stmts[2], "FOREIGN KEY")
	resetSession(t)
}

func TestPostgreSQLConfig(t *testing.T) {
	resetSession(t)
	l := New("postgres")
	sql, err := l.Load(ckmodels.Location{}, ckmodels.Event{}, models.User{}, models.Pet{})
	require.NoError(t, err)
	stmts := strings.Split(sql, "\n")
	require.Contains(t, stmts[0], `CREATE TABLE "events"`)
	require.Contains(t, stmts[2], `CREATE TABLE "locations"`)
	require.Contains(t, stmts[4], `CREATE TABLE "users"`)
	require.Contains(t, stmts[6], `CREATE TABLE "pets"`)
	require.Contains(t, stmts[8], `ALTER TABLE "events" ADD CONSTRAINT "fk_locations_event" FOREIGN KEY ("locationId")`)
	require.Contains(t, stmts[9], `ALTER TABLE "locations" ADD CONSTRAINT "fk_events_location"`)
	require.Contains(t, stmts[10], `ALTER TABLE "pets" ADD CONSTRAINT "fk_users_pets" FOREIGN KEY ("user_id")`)
	resetSession(t)
	l = New("postgres", WithConfig(
		&gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		}))
	sql, err = l.Load(ckmodels.Location{}, ckmodels.Event{})
	require.NoError(t, err)
	stmts = strings.Split(sql, "\n")
	require.Contains(t, stmts[0], `CREATE TABLE "events"`)
	require.Contains(t, stmts[2], `CREATE TABLE "locations"`)
	require.NotContains(t, sql, "FOREIGN KEY")
}

func TestMySQLConfig(t *testing.T) {
	resetSession(t)
	l := New("mysql")
	sql, err := l.Load(ckmodels.Location{}, ckmodels.Event{}, models.User{}, models.Pet{})
	stmts := strings.Split(sql, "\n")
	require.NoError(t, err)
	require.Contains(t, stmts[0], "CREATE TABLE `events`")
	require.Contains(t, stmts[1], "CREATE TABLE `locations`")
	require.Contains(t, stmts[2], "CREATE TABLE `users`")
	require.Contains(t, stmts[3], "CREATE TABLE `pets`")
	require.Contains(t, stmts[4], "ALTER TABLE `events` ADD CONSTRAINT `fk_locations_event`")
	require.Contains(t, stmts[5], "ALTER TABLE `locations` ADD CONSTRAINT `fk_events_location`")
	require.Contains(t, stmts[6], "ALTER TABLE `pets` ADD CONSTRAINT `fk_users_pets`")
	resetSession(t)
	l = New("mysql", WithConfig(
		&gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		},
	))
	sql, err = l.Load(ckmodels.Location{}, ckmodels.Event{})
	require.NoError(t, err)
	stmts = strings.Split(sql, "\n")
	require.Contains(t, stmts[0], "CREATE TABLE `events`")
	require.Contains(t, stmts[1], "CREATE TABLE `locations`")
	require.NotContains(t, sql, "FOREIGN KEY")
}

func resetSession(t *testing.T) {
	sess, ok := recordriver.Session("gorm")
	if ok {
		sess.Statements = nil
	}
}
