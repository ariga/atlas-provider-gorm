-- Create "users" table
CREATE TABLE [users] (
  [id] bigint IDENTITY (1, 1) NOT NULL,
  [created_at] datetimeoffset(7) NULL,
  [updated_at] datetimeoffset(7) NULL,
  [deleted_at] datetimeoffset(7) NULL,
  [name] nvarchar(MAX) COLLATE SQL_Latin1_General_CP1_CI_AS NULL,
  CONSTRAINT [PK_users] PRIMARY KEY CLUSTERED ([id] ASC)
);
-- Create index "idx_users_deleted_at" to table: "users"
CREATE NONCLUSTERED INDEX [idx_users_deleted_at] ON [users] ([deleted_at] ASC);
-- Create "pets" table
CREATE TABLE [pets] (
  [id] bigint IDENTITY (1, 1) NOT NULL,
  [created_at] datetimeoffset(7) NULL,
  [updated_at] datetimeoffset(7) NULL,
  [deleted_at] datetimeoffset(7) NULL,
  [name] nvarchar(MAX) COLLATE SQL_Latin1_General_CP1_CI_AS NULL,
  [user_id] bigint NULL,
  CONSTRAINT [PK_pets] PRIMARY KEY CLUSTERED ([id] ASC),
 
  CONSTRAINT [fk_users_pets] FOREIGN KEY ([user_id]) REFERENCES [users] ([id]) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_pets_deleted_at" to table: "pets"
CREATE NONCLUSTERED INDEX [idx_pets_deleted_at] ON [pets] ([deleted_at] ASC);
