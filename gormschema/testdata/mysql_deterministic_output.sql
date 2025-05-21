-- atlas:pos hobbies[type=table] /internal/testdata/models/user.go:17
-- atlas:pos pets[type=table] /internal/testdata/models/pet.go:11
-- atlas:pos top_pet_owners[type=view] /internal/testdata/models/pet.go:18
-- atlas:pos user_pet_histories[type=table] /internal/testdata/models/pet.go:38
-- atlas:pos users[type=table] /internal/testdata/models/user.go:9
-- atlas:pos working_aged_users[type=view] /internal/testdata/models/user.go:23

CREATE TABLE `hobbies` (`id` bigint unsigned AUTO_INCREMENT,`name` longtext,PRIMARY KEY (`id`));
CREATE TABLE `users` (`id` bigint unsigned AUTO_INCREMENT,`created_at` datetime(3) NULL,`updated_at` datetime(3) NULL,`deleted_at` datetime(3) NULL,`name` longtext,`age` bigint,PRIMARY KEY (`id`),INDEX `idx_users_deleted_at` (`deleted_at`));
CREATE TABLE `user_hobbies` (`user_id` bigint unsigned,`hobby_id` bigint unsigned,PRIMARY KEY (`user_id`,`hobby_id`));
CREATE TABLE `pets` (`id` bigint unsigned AUTO_INCREMENT,`created_at` datetime(3) NULL,`updated_at` datetime(3) NULL,`deleted_at` datetime(3) NULL,`name` longtext,`user_id` bigint unsigned,PRIMARY KEY (`id`),INDEX `idx_pets_deleted_at` (`deleted_at`));
CREATE TABLE `user_pet_histories` (`user_id` bigint unsigned,`pet_id` bigint unsigned,`created_at` datetime(3) NULL,PRIMARY KEY (`user_id`,`pet_id`));
CREATE VIEW top_pet_owners AS SELECT user_id, COUNT(id) AS pet_count FROM pets GROUP BY user_id ORDER BY pet_count DESC LIMIT 10;
CREATE VIEW working_aged_users AS SELECT name, age FROM `users` WHERE age BETWEEN 18 AND 65;
CREATE TRIGGER trg_insert_user_pet_history
AFTER INSERT ON pets
FOR EACH ROW
BEGIN
	INSERT INTO user_pet_histories (user_id, pet_id, created_at)
	VALUES (NEW.user_id, NEW.id, NOW(3));
END;
CREATE TRIGGER trg_adding_heart_on_pet 
BEFORE INSERT ON pets 
FOR EACH ROW
BEGIN
	SET NEW.name = CONCAT(NEW.name, ' <3');
END;
ALTER TABLE `user_hobbies` ADD CONSTRAINT `fk_user_hobbies_hobby` FOREIGN KEY (`hobby_id`) REFERENCES `hobbies`(`id`);
ALTER TABLE `user_hobbies` ADD CONSTRAINT `fk_user_hobbies_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`);
ALTER TABLE `pets` ADD CONSTRAINT `fk_users_pets` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`);
