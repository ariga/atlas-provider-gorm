-- Create "log_user_pet_histories" function
CREATE FUNCTION "public"."log_user_pet_histories" () RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
	INSERT INTO user_pet_histories (user_id, pet_id, created_at)
	VALUES (NEW.user_id, NEW.id, NEW.created_at);
	RETURN NEW;
END;
$$;
-- Create trigger "trg_insert_user_pet_history"
CREATE TRIGGER "trg_insert_user_pet_history" AFTER INSERT ON "public"."pets" FOR EACH ROW EXECUTE FUNCTION "public"."log_user_pet_histories"();
-- Create "add_heart_on_pet" function
CREATE FUNCTION "public"."add_heart_on_pet" () RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
	NEW.name := NEW.name || ' <3';
	RETURN NEW;
END;
$$;
-- Create trigger "trg_adding_heart_on_pet"
CREATE TRIGGER "trg_adding_heart_on_pet" BEFORE INSERT ON "public"."pets" FOR EACH ROW EXECUTE FUNCTION "public"."add_heart_on_pet"();
-- Create "user_pet_histories" table
CREATE TABLE "public"."user_pet_histories" (
  "user_id" bigint NOT NULL,
  "pet_id" bigint NOT NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("user_id", "pet_id")
);
