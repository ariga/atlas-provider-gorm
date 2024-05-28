-- Create trigger "trg_insert_user_pet_history"
CREATE TRIGGER [trg_insert_user_pet_history] ON [pets] AFTER INSERT AS BEGIN
	INSERT INTO user_pet_histories (user_id, pet_id, created_at)
	SELECT 
		inserted.user_id, 
		inserted.id, 
		GETDATE()
	FROM 
		inserted
	WHERE 
		inserted.user_id IS NOT NULL;
END;;
-- Create trigger "trg_adding_heart_on_pet"
CREATE TRIGGER [trg_adding_heart_on_pet] ON [pets] INSTEAD OF INSERT AS BEGIN
	INSERT INTO pets (name, user_id)
	SELECT
		CONCAT(inserted.name, ' ❤️'),
		inserted.user_id
	FROM
		inserted;
END;;
-- Create "user_pet_histories" table
CREATE TABLE [user_pet_histories] (
  [user_id] bigint NOT NULL,
  [pet_id] bigint NOT NULL,
  [created_at] datetimeoffset(7) NULL,
  CONSTRAINT [PK_user_pet_histories] PRIMARY KEY CLUSTERED ([user_id] ASC, [pet_id] ASC)
);
