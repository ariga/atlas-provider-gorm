package gormschema

import (
	"fmt"
	"strings"
	"testing"

	"ariga.io/atlas-provider-gorm/internal/testdata/models"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// `User` belongs to `Company`, `CompanyID` is the foreign key
type User struct {
	gorm.Model
	Name      string
	CompanyID int
	Company   Company
}

type Company struct {
	ID   int
	Name string
}

func Test(t *testing.T) {
	s, err := New("sqlite").Load(&models.User{}, &models.Pet{})
	require.NoError(t, err)
	var ok []int
	fmt.Println(s)
	for i := 0; i < 100; i++ {
		ordered := strings.Index(s, "CREATE TABLE `users`") < strings.Index(s, "CREATE TABLE `pets`")
		if ordered {
			ok = append(ok, i)
		} else {
			t.Logf("not ordered: %d", i)
		}
	}

	t.Logf("ordered: %d", len(ok))
}

func TestGORM(t *testing.T) {
	db, err := gorm.Open(&sqlite.Dialector{
		DSN: "file:testdatabase?mode=memory&cache=shared",
	})
	require.NoError(t, err)
	require.NoError(t, db.Debug().AutoMigrate(&User{}, &Company{}))
}
