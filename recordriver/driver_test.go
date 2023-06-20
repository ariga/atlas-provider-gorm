package recordriver

import (
	"database/sql"
	"database/sql/driver"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDriver(t *testing.T) {
	db, err := sql.Open("recordriver", "t1")
	require.NoError(t, err)
	defer db.Close()
	SetResponse("t1", "select sqlite_version()", &Response{
		Cols: []string{"sqlite_version()"},
		Data: [][]driver.Value{{"3.30.1"}},
	})
	query, err := db.Query("select sqlite_version()")
	require.NoError(t, err)
	defer query.Close()
	for query.Next() {
		var version string
		err = query.Scan(&version)
		require.NoError(t, err)
		require.Equal(t, "3.30.1", version)
	}
	hi, ok := Session("t1")
	require.True(t, ok)
	require.Len(t, hi.Queries, 1)
}

func TestInputs(t *testing.T) {
	db, err := sql.Open("recordriver", "t1")
	require.NoError(t, err)
	defer db.Close()
	_, err = db.Query("select * from t where id = ?", 1)
	require.NoError(t, err)
}
