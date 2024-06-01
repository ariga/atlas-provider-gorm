-- Create "hobbies" table
CREATE TABLE `hobbies` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` longtext NULL,
  PRIMARY KEY (`id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "user_hobbies" table
CREATE TABLE `user_hobbies` (
  `user_id` bigint unsigned NOT NULL,
  `hobby_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`user_id`, `hobby_id`),
  INDEX `fk_user_hobbies_hobby` (`hobby_id`),
  CONSTRAINT `fk_user_hobbies_hobby` FOREIGN KEY (`hobby_id`) REFERENCES `hobbies` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_user_hobbies_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
