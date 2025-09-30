package ds

import (
	"database/sql"
	"time"
)

type ServiceCalc struct {
	ID           uint `gorm:"primaryKey"`
	DateResearch time.Time
	Status       string        `gorm:"type:varchar(15);not null"`
	DateCreate   time.Time     `gorm:"not null"`
	DateForm     sql.NullTime  `gorm:"default:null"`
	DateFinish   sql.NullTime  `gorm:"default:null"`
	CreatorID    int           `gorm:"not null"`
	ModeratorID  sql.NullInt64 `gorm:"default:null"`
	Cost         float64       `gorm:"type: float"`

	Creator   User `gorm:"foreignKey:CreatorID"`
	Moderator User `gorm:"foreignKey:ModeratorID"`
}
