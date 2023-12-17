-- Create "events" table
CREATE TABLE `events` (
  `eventId` text NULL,
  `locationId` text NULL,
  PRIMARY KEY (`eventId`),
  CONSTRAINT `fk_locations_event` FOREIGN KEY (`locationId`) REFERENCES `locations` (`locationId`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_events_location_id" to table: "events"
CREATE UNIQUE INDEX `idx_events_location_id` ON `events` (`locationId`);
-- Create "locations" table
CREATE TABLE `locations` (
  `locationId` text NULL,
  `eventId` text NULL,
  PRIMARY KEY (`locationId`),
  CONSTRAINT `fk_events_location` FOREIGN KEY (`eventId`) REFERENCES `events` (`eventId`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_locations_event_id" to table: "locations"
CREATE UNIQUE INDEX `idx_locations_event_id` ON `locations` (`eventId`);
