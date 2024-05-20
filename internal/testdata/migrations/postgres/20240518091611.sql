-- Create "top_pet_owners" view
CREATE VIEW "public"."top_pet_owners" (
  "user_id",
  "pet_count"
) AS SELECT pets.user_id,
    count(pets.id) AS pet_count
   FROM pets
  GROUP BY pets.user_id
  ORDER BY (count(pets.id)) DESC
 LIMIT 10;
-- Create "working_aged_users" view
CREATE VIEW "public"."working_aged_users" (
  "name",
  "age"
) AS SELECT users.name,
    users.age
   FROM users
  WHERE ((users.age >= 18) AND (users.age <= 65));
