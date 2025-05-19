package main

import (
	"bytes"
	"os"
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
}

func TestDeterministicOutput(t *testing.T) {
	expected, err := os.ReadFile("./gormschema/testdata/mysql_deterministic_output")
	require.NoError(t, err)
	cmd := &LoadCmd{
		Path:    "./internal/testdata/models",
		Dialect: "mysql",
	}
	for i := 0; i < 10; i++ {
		var buf bytes.Buffer
		cmd.out = &buf
		err := cmd.Run()
		require.NoError(t, err)
		actual := bytes.TrimSuffix(buf.Bytes(), []byte("\n"))
		require.Equal(t, string(expected), string(actual))
	}
}

func TestCustomizeTablesLoad(t *testing.T) {
	var buf bytes.Buffer
	cmd := &LoadCmd{
		Path:    "./internal/testdata/customjointable",
		Dialect: "mysql",
		out:     &buf,
	}
	err := cmd.Run()
	require.NoError(t, err)
	expected, err := os.ReadFile("./gormschema/testdata/mysql_custom_join_table")
	require.NoError(t, err)
	actual := bytes.TrimSuffix(buf.Bytes(), []byte("\n"))
	require.Equal(t, string(expected), string(actual))
}

func TestTaggedModels(t *testing.T) {
	var buf bytes.Buffer
	cmd := &LoadCmd{
		Path:      "./internal/testdata/taggedmodels",
		Dialect:   "mysql",
		BuildTags: "tag",
		out:       &buf,
	}
	err := cmd.Run()
	require.NoError(t, err)
	require.Contains(t, buf.String(), "CREATE TABLE `non_tagged_models`")
	require.Contains(t, buf.String(), "CREATE TABLE `tagged_models`")
}

func TestNonTaggedModels(t *testing.T) {
	var buf bytes.Buffer
	cmd := &LoadCmd{
		Path:    "./internal/testdata/taggedmodels",
		Dialect: "mysql",
		out:     &buf,
	}
	err := cmd.Run()
	require.NoError(t, err)
	require.Contains(t, buf.String(), "CREATE TABLE `non_tagged_models`")
	require.NotContains(t, buf.String(), "CREATE TABLE `tagged_models`")
}
