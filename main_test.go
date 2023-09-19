package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	testCases := []struct {
		dialect                              string
		disableMigrationForeignKeyConstraint bool
	}{
		{dialect: "mysql", disableMigrationForeignKeyConstraint: false},
		{dialect: "sqlite", disableMigrationForeignKeyConstraint: false},
		{dialect: "postgres", disableMigrationForeignKeyConstraint: false},
		{dialect: "mysql", disableMigrationForeignKeyConstraint: true},
		{dialect: "sqlite", disableMigrationForeignKeyConstraint: true},
		{dialect: "postgres", disableMigrationForeignKeyConstraint: true},
	}

	for _, tc := range testCases {
		t.Run(tc.dialect, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := &LoadCmd{
				Path:                                 "./internal/testdata/models",
				Dialect:                              tc.dialect,
				out:                                  &buf,
				DisableMigrationForeignKeyConstraint: tc.disableMigrationForeignKeyConstraint,
			}
			err := cmd.Run()
			require.NoError(t, err)

			assertLoadOutput(t, buf.String(), tc.disableMigrationForeignKeyConstraint)
		})
	}
}

func assertLoadOutput(t *testing.T, output string, disableMigrationForeignKeyConstraint bool) {
	require.Contains(t, output, "CREATE TABLE")
	require.Contains(t, output, "pets")
	require.Contains(t, output, "users")
	require.NotContains(t, output, "toys") // Struct without GORM annotations

	if disableMigrationForeignKeyConstraint {
		require.NotContains(t, output, "CONSTRAINT")
		require.NotContains(t, output, "FOREIGN KEY")
	} else {
		require.Contains(t, output, "CONSTRAINT")
		require.Contains(t, output, "FOREIGN KEY")
	}
}
