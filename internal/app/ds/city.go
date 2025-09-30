package ds

type City struct {
	ID                int     `gorm:"primaryKey"`
	IsDelete          bool    `gorm:"type:boolean not null;default:false"`
	Image             string  `gorm:"type: varchar(32)"`
	Name              string  `gorm:"type:varchar(25);not null"`
	Central           bool    `gorm:"type:boolean not null;default:false"`
	Description       string  `gorm:"type: varchar(512)"`
	Coordinates_long  float64 `gorm:"type: float"`
	Coordinates_width float64 `gorm:"type: float"`
}
