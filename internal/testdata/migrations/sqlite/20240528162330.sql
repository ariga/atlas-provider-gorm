-- Create trigger "trg_insert_user_pet_history"
CREATE TRIGGER `trg_insert_user_pet_history` AFTER INSERT ON `pets` FOR EACH ROW BEGIN
	INSERT INTO user_pet_histories (user_id, pet_id, created_at)
	VALUES (NEW.user_id, NEW.id, datetime('now'));
END;
-- Create trigger "trg_adding_heart_on_pet"
CREATE TRIGGER `trg_adding_heart_on_pet` BEFORE INSERT ON `pets` FOR EACH ROW BEGIN
	UPDATE pets SET name = name || ' <3' WHERE id = NEW.id;
-- Create "user_pet_histories" table
CREATE TABLE `user_pet_histories` (
  `user_id` integer NULL,
  `pet_id` integer NULL,
  `created_at` datetime NULL,
  PRIMARY KEY (`user_id`, `pet_id`)
);
