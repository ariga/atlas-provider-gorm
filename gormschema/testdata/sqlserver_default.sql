CREATE TABLE "events" ("eventId" nvarchar(191),"locationId" nvarchar(191),PRIMARY KEY ("eventId"))
GO
CREATE UNIQUE INDEX "idx_events_location_id" ON "events"("locationId")
GO
CREATE TABLE "locations" ("locationId" nvarchar(191),"eventId" nvarchar(191),PRIMARY KEY ("locationId"))
GO
CREATE UNIQUE INDEX "idx_locations_event_id" ON "locations"("eventId")
GO
CREATE TABLE "user_pet_histories" ("user_id" bigint,"pet_id" bigint,"created_at" datetimeoffset,PRIMARY KEY ("user_id","pet_id"))
GO
CREATE TABLE "users" ("id" bigint IDENTITY(1,1),"created_at" datetimeoffset,"updated_at" datetimeoffset,"deleted_at" datetimeoffset,"name" nvarchar(MAX),"age" bigint,PRIMARY KEY ("id"))
GO
CREATE INDEX "idx_users_deleted_at" ON "users"("deleted_at")
GO
CREATE TABLE "hobbies" ("id" bigint IDENTITY(1,1),"name" nvarchar(MAX),PRIMARY KEY ("id"))
GO
CREATE TABLE "user_hobbies" ("hobby_id" bigint,"user_id" bigint,PRIMARY KEY ("hobby_id","user_id"))
GO
CREATE TABLE "pets" ("id" bigint IDENTITY(1,1),"created_at" datetimeoffset,"updated_at" datetimeoffset,"deleted_at" datetimeoffset,"name" nvarchar(MAX),"user_id" bigint,PRIMARY KEY ("id"))
GO
CREATE INDEX "idx_pets_deleted_at" ON "pets"("deleted_at")
GO
CREATE VIEW working_aged_users AS SELECT name, age FROM "users" WHERE age BETWEEN 18 AND 65
GO
CREATE VIEW top_pet_owners AS SELECT user_id, COUNT(id) AS pet_count FROM pets GROUP BY user_id ORDER BY pet_count DESC OFFSET 0 ROWS FETCH NEXT 10 ROWS ONLY
GO
CREATE TRIGGER trg_insert_user_pet_history
ON pets
AFTER INSERT
AS
BEGIN
	INSERT INTO user_pet_histories (user_id, pet_id, created_at)
	SELECT 
		inserted.user_id, 
		inserted.id, 
		GETDATE()
	FROM 
		inserted
	WHERE 
		inserted.user_id IS NOT NULL;
END
GO
CREATE TRIGGER trg_adding_heart_on_pet
ON pets
INSTEAD OF INSERT
AS
BEGIN
	INSERT INTO pets (name, user_id)
	SELECT
		CONCAT(inserted.name, ' <3'),
		inserted.user_id
	FROM
		inserted;
END
GO
ALTER TABLE "events" ADD CONSTRAINT "fk_locations_event" FOREIGN KEY ("locationId") REFERENCES "locations"("locationId")
GO
ALTER TABLE "locations" ADD CONSTRAINT "fk_events_location" FOREIGN KEY ("eventId") REFERENCES "events"("eventId")
GO
ALTER TABLE "user_hobbies" ADD CONSTRAINT "fk_user_hobbies_hobby" FOREIGN KEY ("hobby_id") REFERENCES "hobbies"("id")
GO
ALTER TABLE "user_hobbies" ADD CONSTRAINT "fk_user_hobbies_user" FOREIGN KEY ("user_id") REFERENCES "users"("id")
GO
ALTER TABLE "pets" ADD CONSTRAINT "fk_users_pets" FOREIGN KEY ("user_id") REFERENCES "users"("id")
GO
