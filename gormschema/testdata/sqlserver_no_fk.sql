CREATE TABLE "events" ("eventId" nvarchar(191),"locationId" nvarchar(191),PRIMARY KEY ("eventId"))
GO
CREATE UNIQUE INDEX "idx_events_location_id" ON "events"("locationId")
GO
CREATE TABLE "locations" ("locationId" nvarchar(191),"eventId" nvarchar(191),PRIMARY KEY ("locationId"))
GO
CREATE UNIQUE INDEX "idx_locations_event_id" ON "locations"("eventId")
GO
