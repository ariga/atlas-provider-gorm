-- Create trigger "trg_adding_heart_on_pet"
CREATE TRIGGER `trg_adding_heart_on_pet` BEFORE INSERT ON `pets` FOR EACH ROW BEGIN
	SET NEW.name = CONCAT(NEW.name, ' ❤️');
END;
-- Create trigger "trg_insert_user_pet_history"
CREATE TRIGGER `trg_insert_user_pet_history` AFTER INSERT ON `pets` FOR EACH ROW BEGIN
	INSERT INTO user_pet_histories (user_id, pet_id, created_at)
	VALUES (NEW.user_id, NEW.id, NOW(3));
END;
-- Create "user_pet_histories" table
CREATE TABLE `user_pet_histories` (
  `user_id` bigint unsigned NOT NULL,
  `pet_id` bigint unsigned NOT NULL,
  `created_at` datetime(3) NULL,
  PRIMARY KEY (`user_id`, `pet_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
