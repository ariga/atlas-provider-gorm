-- Create "top_pet_owners" view
CREATE VIEW `top_pet_owners` (
  `user_id`,
  `pet_count`
) AS SELECT user_id, COUNT(id) AS pet_count FROM pets GROUP BY user_id ORDER BY pet_count DESC LIMIT 10;
-- Create "working_aged_users" view
CREATE VIEW `working_aged_users` (
  `name`,
  `age`
) AS SELECT name, age FROM `users` WHERE age BETWEEN 18 AND 65;
