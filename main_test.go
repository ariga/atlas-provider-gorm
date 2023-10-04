package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	testCases := []struct {
		dialect            string
		disableForeignKeys bool
	}{
		{dialect: "mysql", disableForeignKeys: false},
		{dialect: "sqlite", disableForeignKeys: false},
		{dialect: "postgres", disableForeignKeys: false},
		{dialect: "mysql", disableForeignKeys: true},
		{dialect: "sqlite", disableForeignKeys: true},
		{dialect: "postgres", disableForeignKeys: true},
	}

	for _, tc := range testCases {
		t.Run(tc.dialect, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := &LoadCmd{
				Path:               "./internal/testdata/models",
				Dialect:            tc.dialect,
				out:                &buf,
				DisableForeignKeys: tc.disableForeignKeys,
			}
			err := cmd.Run()
			require.NoError(t, err)

			assertLoadOutput(t, buf.String(), tc.disableForeignKeys)
		})
	}
}

func assertLoadOutput(t *testing.T, output string, disableForeignKeys bool) {
	require.Contains(t, output, "CREATE TABLE")
	require.Contains(t, output, "pets")
	require.Contains(t, output, "users")
	require.NotContains(t, output, "toys") // Struct without GORM annotations

	if disableForeignKeys {
		require.NotContains(t, output, "CONSTRAINT")
		require.NotContains(t, output, "FOREIGN KEY")
	} else {
		require.Contains(t, output, "CONSTRAINT")
		require.Contains(t, output, "FOREIGN KEY")
	}
}
