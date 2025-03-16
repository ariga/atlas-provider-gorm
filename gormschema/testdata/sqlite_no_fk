CREATE TABLE `user_pet_histories` (`user_id` integer,`pet_id` integer,`created_at` datetime,PRIMARY KEY (`user_id`,`pet_id`));
CREATE TABLE `users` (`id` integer PRIMARY KEY AUTOINCREMENT,`created_at` datetime,`updated_at` datetime,`deleted_at` datetime,`name` text,`age` integer);
CREATE INDEX `idx_users_deleted_at` ON `users`(`deleted_at`);
CREATE TABLE `pets` (`id` integer PRIMARY KEY AUTOINCREMENT,`created_at` datetime,`updated_at` datetime,`deleted_at` datetime,`name` text,`user_id` integer);
CREATE INDEX `idx_pets_deleted_at` ON `pets`(`deleted_at`);
CREATE TABLE `hobbies` (`id` integer PRIMARY KEY AUTOINCREMENT,`name` text);
CREATE TABLE `user_hobbies` (`hobby_id` integer,`user_id` integer,PRIMARY KEY (`hobby_id`,`user_id`));
CREATE TRIGGER trg_insert_user_pet_history
AFTER INSERT ON pets
BEGIN
	INSERT INTO user_pet_histories (user_id, pet_id, created_at)
	VALUES (NEW.user_id, NEW.id, datetime('now'));
END;
CREATE TRIGGER trg_adding_heart_on_pet
BEFORE INSERT ON pets
BEGIN
	UPDATE pets SET name = name || ' <3' WHERE id = NEW.id;
END;
