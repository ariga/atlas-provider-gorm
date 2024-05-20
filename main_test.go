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
}

func TestDeterministicOutput(t *testing.T) {
	expected := "CREATE TABLE `users` (`id` integer,`created_at` datetime,`updated_at` datetime,`deleted_at` datetime,`name` text,`age` integer,PRIMARY KEY (`id`));\n" +
		"CREATE INDEX `idx_users_deleted_at` ON `users`(`deleted_at`);\n" +
		"CREATE TABLE `pets` (`id` integer,`created_at` datetime,`updated_at` datetime,`deleted_at` datetime,`name` text,`user_id` integer,PRIMARY KEY (`id`),CONSTRAINT `fk_users_pets` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`));\nCREATE INDEX `idx_pets_deleted_at` ON `pets`(`deleted_at`);\n" +
		"CREATE VIEW top_pet_owners AS SELECT user_id, COUNT(id) AS pet_count FROM pets GROUP BY user_id ORDER BY pet_count DESC LIMIT 10;\n" +
		"CREATE VIEW working_aged_users AS SELECT name, age FROM `users` WHERE age BETWEEN 18 AND 65;\n\n"
	cmd := &LoadCmd{
		Path:    "./internal/testdata/models",
		Dialect: "sqlite",
	}
	for i := 0; i < 10; i++ {
		var buf bytes.Buffer
		cmd.out = &buf
		err := cmd.Run()
		require.NoError(t, err)
		require.Equal(t, expected, buf.String())
	}
}
