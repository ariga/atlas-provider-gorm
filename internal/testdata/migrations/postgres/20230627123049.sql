-- Create "users" table
CREATE TABLE "public"."users" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "name" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_users_deleted_at" to table: "users"
CREATE INDEX "idx_users_deleted_at" ON "public"."users" ("deleted_at");
-- Create "pets" table
CREATE TABLE "public"."pets" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "name" text NULL,
  "user_id" bigint NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_users_pets" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_pets_deleted_at" to table: "pets"
CREATE INDEX "idx_pets_deleted_at" ON "public"."pets" ("deleted_at");
