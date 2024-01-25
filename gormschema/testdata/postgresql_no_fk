CREATE TABLE "events" ("eventId" varchar(191),"locationId" varchar(191),PRIMARY KEY ("eventId"));
CREATE UNIQUE INDEX IF NOT EXISTS "idx_events_location_id" ON "events" ("locationId");
CREATE TABLE "locations" ("locationId" varchar(191),"eventId" varchar(191),PRIMARY KEY ("locationId"));
CREATE UNIQUE INDEX IF NOT EXISTS "idx_locations_event_id" ON "locations" ("eventId");
