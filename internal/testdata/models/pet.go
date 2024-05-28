package models

import (
	"time"

	"gorm.io/gorm"

	"ariga.io/atlas-provider-gorm/gormschema"
)

type Pet struct {
	gorm.Model
	Name   string
	User   User
	UserID uint
}

type TopPetOwner struct {
	Name     string
	PetCount int
}

func (TopPetOwner) ViewDef(dialect string) []gormschema.ViewOption {
	var stmt string
	switch dialect {
	case "mysql":
		stmt = "CREATE VIEW top_pet_owners AS SELECT user_id, COUNT(id) AS pet_count FROM pets GROUP BY user_id ORDER BY pet_count DESC LIMIT 10"
	case "postgres":
		stmt = "CREATE VIEW top_pet_owners AS SELECT user_id, COUNT(id) AS pet_count FROM pets GROUP BY user_id ORDER BY pet_count DESC LIMIT 10"
	case "sqlite":
		stmt = "CREATE VIEW top_pet_owners AS SELECT user_id, COUNT(id) AS pet_count FROM pets GROUP BY user_id ORDER BY pet_count DESC LIMIT 10"
	case "sqlserver":
		stmt = "CREATE VIEW top_pet_owners AS SELECT user_id, COUNT(id) AS pet_count FROM pets GROUP BY user_id ORDER BY pet_count DESC OFFSET 0 ROWS FETCH NEXT 10 ROWS ONLY"
	}
	return []gormschema.ViewOption{gormschema.CreateStmt(stmt)}
}

type UserPetHistory struct {
	UserID    uint `gorm:"primaryKey"`
	PetID     uint `gorm:"primaryKey"`
	CreatedAt time.Time
}

func (Pet) Triggers(dialect string) [][]gormschema.TriggerOption {
	var stmt1, stmt2 string
	switch dialect {
	case "mysql":
		stmt1 = `CREATE TRIGGER trg_insert_user_pet_history
AFTER INSERT ON pets
FOR EACH ROW
BEGIN
	INSERT INTO user_pet_histories (user_id, pet_id, created_at)
	VALUES (NEW.user_id, NEW.id, NOW(3));
END`
		stmt2 = `CREATE TRIGGER trg_adding_heart_on_pet 
BEFORE INSERT ON pets 
FOR EACH ROW
BEGIN
	SET NEW.name = CONCAT(NEW.name, ' ❤️');
END`
	case "sqlite":
		stmt1 = `CREATE TRIGGER trg_insert_user_pet_history
AFTER INSERT ON pets
BEGIN
	INSERT INTO user_pet_histories (user_id, pet_id, created_at)
	VALUES (NEW.user_id, NEW.id, datetime('now'));
END`
		stmt2 = `CREATE TRIGGER trg_adding_heart_on_pet
BEFORE INSERT ON pets
BEGIN
	UPDATE pets SET name = name || ' ❤️' WHERE id = NEW.id;
END`
	case "postgres":
		stmt1 = `CREATE OR REPLACE FUNCTION log_user_pet_histories()
RETURNS TRIGGER AS $$
BEGIN
	INSERT INTO user_pet_histories (user_id, pet_id, created_at)
	VALUES (NEW.user_id, NEW.id, NEW.created_at);
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_insert_user_pet_history
AFTER INSERT ON pets
FOR EACH ROW
EXECUTE FUNCTION log_user_pet_histories();`
		stmt2 = `CREATE OR REPLACE FUNCTION add_heart_on_pet()
RETURNS TRIGGER AS $$
BEGIN
	NEW.name := NEW.name || ' ❤️';
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_adding_heart_on_pet
BEFORE INSERT ON pets
FOR EACH ROW
EXECUTE FUNCTION add_heart_on_pet();`
	case "sqlserver":
		stmt1 = `CREATE TRIGGER trg_insert_user_pet_history
ON pets
AFTER INSERT
AS
BEGIN
	INSERT INTO user_pet_histories (user_id, pet_id, created_at)
	SELECT 
		inserted.user_id, 
		inserted.id, 
		GETDATE()
	FROM 
		inserted
	WHERE 
		inserted.user_id IS NOT NULL;
END`
		stmt2 = `CREATE TRIGGER trg_adding_heart_on_pet
ON pets
INSTEAD OF INSERT
AS
BEGIN
	INSERT INTO pets (name, user_id)
	SELECT
		CONCAT(inserted.name, ' ❤️'),
		inserted.user_id
	FROM
		inserted;
END`
	}

	return [][]gormschema.TriggerOption{
		[]gormschema.TriggerOption{gormschema.CreateStmt(stmt1)},
		[]gormschema.TriggerOption{gormschema.CreateStmt(stmt2)},
	}
}
