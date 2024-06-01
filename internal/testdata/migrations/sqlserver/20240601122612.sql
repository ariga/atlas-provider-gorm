-- Create "hobbies" table
CREATE TABLE [hobbies] (
  [id] bigint IDENTITY (1, 1) NOT NULL,
  [name] nvarchar(MAX) COLLATE SQL_Latin1_General_CP1_CI_AS NULL,
  CONSTRAINT [PK_hobbies] PRIMARY KEY CLUSTERED ([id] ASC)
);
-- Create "user_hobbies" table
CREATE TABLE [user_hobbies] (
  [user_id] bigint NOT NULL,
  [hobby_id] bigint NOT NULL,
  CONSTRAINT [PK_user_hobbies] PRIMARY KEY CLUSTERED ([user_id] ASC, [hobby_id] ASC),
 
  CONSTRAINT [fk_user_hobbies_hobby] FOREIGN KEY ([hobby_id]) REFERENCES [hobbies] ([id]) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT [fk_user_hobbies_user] FOREIGN KEY ([user_id]) REFERENCES [users] ([id]) ON UPDATE NO ACTION ON DELETE NO ACTION
);
