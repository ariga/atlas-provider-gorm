package circularfks

type Location struct {
	LocationID string `gorm:"primaryKey;column:locationId;size:191"`
	EventID    string `gorm:"uniqueIndex;column:eventId;size:191"`
	Event      *Event `gorm:"foreignKey:locationId;references:locationId;OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type Event struct {
	EventID    string    `gorm:"primaryKey;column:eventId;size:191"`
	LocationID string    `gorm:"uniqueIndex;column:locationId;size:191"`
	Location   *Location `gorm:"foreignKey:eventId;references:eventId;OnUpdate:CASCADE,OnDelete:CASCADE"`
}
