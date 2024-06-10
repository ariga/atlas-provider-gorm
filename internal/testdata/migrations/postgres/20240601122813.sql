-- Create "hobbies" table
CREATE TABLE "public"."hobbies" (
  "id" bigserial NOT NULL,
  "name" text NULL,
  PRIMARY KEY ("id")
);
-- Create "user_hobbies" table
CREATE TABLE "public"."user_hobbies" (
  "user_id" bigint NOT NULL,
  "hobby_id" bigint NOT NULL,
  PRIMARY KEY ("user_id", "hobby_id"),
  CONSTRAINT "fk_user_hobbies_hobby" FOREIGN KEY ("hobby_id") REFERENCES "public"."hobbies" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_user_hobbies_user" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
