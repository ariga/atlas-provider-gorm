-- Create "top_pet_owners" view
CREATE VIEW `top_pet_owners` (
  `user_id`,
  `pet_count`
) AS select `pets`.`user_id` AS `user_id`,count(`pets`.`id`) AS `pet_count` from `pets` group by `pets`.`user_id` order by `pet_count` desc limit 10;
-- Create "working_aged_users" view
CREATE VIEW `working_aged_users` (
  `name`,
  `age`
) AS select `users`.`name` AS `name`,`users`.`age` AS `age` from `users` where (`users`.`age` between 18 and 65);
