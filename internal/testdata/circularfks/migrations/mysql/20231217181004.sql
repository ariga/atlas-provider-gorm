-- Create "events" table
CREATE TABLE `events` (
  `eventId` varchar(191) NOT NULL,
  `locationId` varchar(191) NULL,
  PRIMARY KEY (`eventId`),
  UNIQUE INDEX `idx_events_location_id` (`locationId`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "locations" table
CREATE TABLE `locations` (
  `locationId` varchar(191) NOT NULL,
  `eventId` varchar(191) NULL,
  PRIMARY KEY (`locationId`),
  UNIQUE INDEX `idx_locations_event_id` (`eventId`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Modify "events" table
ALTER TABLE `events` ADD CONSTRAINT `fk_locations_event` FOREIGN KEY (`locationId`) REFERENCES `locations` (`locationId`) ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "locations" table
ALTER TABLE `locations` ADD CONSTRAINT `fk_events_location` FOREIGN KEY (`eventId`) REFERENCES `events` (`eventId`) ON UPDATE NO ACTION ON DELETE NO ACTION;
