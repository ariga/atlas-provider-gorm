package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	for _, dialect := range []string{"mysql", "sqlite", "postgres", "sqlserver"} {
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
	expected, err := os.ReadFile("./gormschema/testdata/mysql_deterministic_output.sql")
	require.NoError(t, err)
	cmd := &LoadCmd{Path: "./internal/testdata/models", Dialect: "mysql"}
	cwd, err := os.Getwd()
	require.NoError(t, err)
	for range 10 {
		var buf bytes.Buffer
		cmd.out = &buf
		err := cmd.Run()
		require.NoError(t, err)
		actual := bytes.TrimSuffix(buf.Bytes(), []byte("\n"))
		require.Equal(t, string(expected), strings.ReplaceAll(string(actual), cwd, ""))
	}
}

func TestCustomizeTablesLoad(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)
	var buf bytes.Buffer
	cmd := &LoadCmd{
		Path:    "./internal/testdata/customjointable",
		Dialect: "mysql",
		out:     &buf,
	}
	require.NoError(t, cmd.Run())
	expected, err := os.ReadFile("./gormschema/testdata/mysql_custom_join_table.sql")
	require.NoError(t, err)
	actual := bytes.TrimSuffix(buf.Bytes(), []byte("\n"))
	require.Equal(t, string(expected), strings.ReplaceAll(string(actual), cwd, ""))
}

func TestBuildTags(t *testing.T) {
	var buf bytes.Buffer
	cmd := &LoadCmd{
		Path:      "./internal/testdata/buildtags",
		Dialect:   "mysql",
		BuildTags: "buildtag",
		out:       &buf,
	}
	err := cmd.Run()
	require.NoError(t, err)
	require.Contains(t, buf.String(), "CREATE TABLE `untagged_models`")
	require.Contains(t, buf.String(), "CREATE TABLE `tagged_models`")
}

func TestNonBuildTags(t *testing.T) {
	var buf bytes.Buffer
	cmd := &LoadCmd{
		Path:    "./internal/testdata/buildtags",
		Dialect: "mysql",
		out:     &buf,
	}
	err := cmd.Run()
	require.NoError(t, err)
	require.Contains(t, buf.String(), "CREATE TABLE `untagged_models`")
	require.NotContains(t, buf.String(), "CREATE TABLE `tagged_models`")
}
