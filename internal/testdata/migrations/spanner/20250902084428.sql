-- Create "user_pet_histories" table
CREATE TABLE `user_pet_histories` (
  `user_id` INT64,
  `pet_id` INT64,
  `created_at` TIMESTAMP
) PRIMARY KEY (`user_id`, `pet_id`);
-- Create "users" table
CREATE TABLE `users` (
  `id` INT64,
  `created_at` TIMESTAMP,
  `updated_at` TIMESTAMP,
  `deleted_at` TIMESTAMP,
  `name` STRING(2621440),
  `age` INT64
) PRIMARY KEY (`id`);
-- Create "idx_users_deleted_at" index
CREATE INDEX `idx_users_deleted_at` ON `users` (`deleted_at`);
-- Create "pets" table
CREATE TABLE `pets` (
  `id` INT64,
  `created_at` TIMESTAMP,
  `updated_at` TIMESTAMP,
  `deleted_at` TIMESTAMP,
  `name` STRING(2621440),
  `user_id` INT64,
  CONSTRAINT `fk_users_pets` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE NO ACTION
) PRIMARY KEY (`id`);
-- Create "idx_pets_deleted_at" index
CREATE INDEX `idx_pets_deleted_at` ON `pets` (`deleted_at`);
-- Create "hobbies" table
CREATE TABLE `hobbies` (
  `id` INT64,
  `name` STRING(2621440)
) PRIMARY KEY (`id`);
-- Create "user_hobbies" table
CREATE TABLE `user_hobbies` (
  `user_id` INT64,
  `hobby_id` INT64,
  CONSTRAINT `fk_user_hobbies_hobby` FOREIGN KEY (`hobby_id`) REFERENCES `hobbies` (`id`) ON DELETE NO ACTION,
  CONSTRAINT `fk_user_hobbies_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE NO ACTION
) PRIMARY KEY (`user_id`, `hobby_id`);
-- Create view named "working_aged_users"
CREATE OR REPLACE VIEW `working_aged_users` SQL SECURITY INVOKER AS SELECT u.name, u.age
FROM users AS u
WHERE u.age BETWEEN 18 AND 65;
