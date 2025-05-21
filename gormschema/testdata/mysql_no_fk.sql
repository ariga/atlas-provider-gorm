CREATE TABLE `events` (`eventId` varchar(191),`locationId` varchar(191),PRIMARY KEY (`eventId`),UNIQUE INDEX `idx_events_location_id` (`locationId`));
CREATE TABLE `locations` (`locationId` varchar(191),`eventId` varchar(191),PRIMARY KEY (`locationId`),UNIQUE INDEX `idx_locations_event_id` (`eventId`));
