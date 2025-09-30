package ds

type CityInfo struct {
	ID                int     `gorm:"primaryKey"`
	Image             string  `gorm:"type: varchar(32)"`
	Name              string  `gorm:"type:varchar(25);not null"`
	Central           bool    `gorm:"type:boolean not null;default:false"`
	Coordinates_long  float64 `gorm:"type: float"`
	Coordinates_width float64 `gorm:"type: float"`
}
