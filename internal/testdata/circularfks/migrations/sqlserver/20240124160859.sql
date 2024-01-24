-- Create "events" table
CREATE TABLE [events] (
  [eventId] nvarchar(191) COLLATE SQL_Latin1_General_CP1_CI_AS NOT NULL,
  [locationId] nvarchar(191) COLLATE SQL_Latin1_General_CP1_CI_AS NULL,
  CONSTRAINT [PK_events] PRIMARY KEY CLUSTERED ([eventId] ASC)
);
-- Create index "idx_events_location_id" to table: "events"
CREATE UNIQUE NONCLUSTERED INDEX [idx_events_location_id] ON [events] ([locationId] ASC);
-- Create "locations" table
CREATE TABLE [locations] (
  [locationId] nvarchar(191) COLLATE SQL_Latin1_General_CP1_CI_AS NOT NULL,
  [eventId] nvarchar(191) COLLATE SQL_Latin1_General_CP1_CI_AS NULL,
  CONSTRAINT [PK_locations] PRIMARY KEY CLUSTERED ([locationId] ASC)
);
-- Create index "idx_locations_event_id" to table: "locations"
CREATE UNIQUE NONCLUSTERED INDEX [idx_locations_event_id] ON [locations] ([eventId] ASC);
-- Modify "events" table
ALTER TABLE [events] ADD
 CONSTRAINT [fk_locations_event] FOREIGN KEY ([locationId]) REFERENCES [locations] ([locationId]) ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "locations" table
ALTER TABLE [locations] ADD
 CONSTRAINT [fk_events_location] FOREIGN KEY ([eventId]) REFERENCES [events] ([eventId]) ON UPDATE NO ACTION ON DELETE NO ACTION;
