-- Create "events" table
CREATE TABLE "public"."events" (
  "eventId" character varying(191) NOT NULL,
  "locationId" character varying(191) NULL,
  PRIMARY KEY ("eventId")
);
-- Create index "idx_events_location_id" to table: "events"
CREATE UNIQUE INDEX "idx_events_location_id" ON "public"."events" ("locationId");
-- Create "locations" table
CREATE TABLE "public"."locations" (
  "locationId" character varying(191) NOT NULL,
  "eventId" character varying(191) NULL,
  PRIMARY KEY ("locationId")
);
-- Create index "idx_locations_event_id" to table: "locations"
CREATE UNIQUE INDEX "idx_locations_event_id" ON "public"."locations" ("eventId");
-- Modify "events" table
ALTER TABLE "public"."events" ADD
 CONSTRAINT "fk_locations_event" FOREIGN KEY ("locationId") REFERENCES "public"."locations" ("locationId") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "locations" table
ALTER TABLE "public"."locations" ADD
 CONSTRAINT "fk_events_location" FOREIGN KEY ("eventId") REFERENCES "public"."events" ("eventId") ON UPDATE NO ACTION ON DELETE NO ACTION;
