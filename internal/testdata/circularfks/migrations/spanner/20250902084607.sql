-- Create "events" table
CREATE TABLE `events` (
  `eventId` STRING(191),
  `locationId` STRING(191)
) PRIMARY KEY (`eventId`);
-- Create "idx_events_location_id" index
CREATE UNIQUE INDEX `idx_events_location_id` ON `events` (`locationId`);
-- Create "locations" table
CREATE TABLE `locations` (
  `locationId` STRING(191),
  `eventId` STRING(191)
) PRIMARY KEY (`locationId`);
-- Create "idx_locations_event_id" index
CREATE UNIQUE INDEX `idx_locations_event_id` ON `locations` (`eventId`);
-- Add foreign key "fk_locations_event" to table "events"
ALTER TABLE `events` ADD CONSTRAINT `fk_locations_event` FOREIGN KEY (`locationId`) REFERENCES `locations` (`locationId`) ON DELETE NO ACTION;
-- Add foreign key "fk_events_location" to table "locations"
ALTER TABLE `locations` ADD CONSTRAINT `fk_events_location` FOREIGN KEY (`eventId`) REFERENCES `events` (`eventId`) ON DELETE NO ACTION;
