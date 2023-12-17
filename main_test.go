package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	for _, dialect := range []string{"mysql", "sqlite", "postgres"} {
		t.Run(dialect, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := &LoadCmd{
				Path:    "./internal/testdata/models",
				Dialect: dialect,
				out:     &buf,
			}
			err := cmd.Run()
			require.NoError(t, err)
			require.Contains(t, buf.String(), "CREATE TABLE")
			require.Contains(t, buf.String(), "pets")
			require.Contains(t, buf.String(), "users")
			require.NotContains(t, buf.String(), "toys") // Struct without GORM annotations.
		})
	}
	for _, dialect := range []string{"mysql", "sqlite", "postgres"} {
		t.Run(dialect, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := &LoadCmd{
				Path:                                "./internal/testdata/models",
				Dialect:                             dialect,
				CreateConstraintsAfterTableCreation: true,
				out:                                 &buf,
			}
			err := cmd.Run()
			if dialect == "sqlite" {
				// sqlite does not support alter columns
				require.Errorf(t, err, "invalid DDL")
			} else {
				require.NoError(t, err)
				require.Contains(t, buf.String(), "ALTER TABLE")
			}
		})
	}
}
