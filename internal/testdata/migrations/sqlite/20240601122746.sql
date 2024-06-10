-- Create "hobbies" table
CREATE TABLE `hobbies` (
  `id` integer NULL,
  `name` text NULL,
  PRIMARY KEY (`id`)
);
-- Create "user_hobbies" table
CREATE TABLE `user_hobbies` (
  `user_id` integer NULL,
  `hobby_id` integer NULL,
  PRIMARY KEY (`user_id`, `hobby_id`),
  CONSTRAINT `fk_user_hobbies_hobby` FOREIGN KEY (`hobby_id`) REFERENCES `hobbies` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `fk_user_hobbies_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
