package ds

type CityForService struct {
	ID uint `gorm:"primaryKey"`

	ServiceCalcID uint `gorm:"not null;uniqueIndex:idx_research_chat"`
	CityID        uint `gorm:"not null;uniqueIndex:idx_research_chat"`

	ServiceCalc ServiceCalc `gorm:"foreignKey:ServiceCalcID"`
	City        City        `gorm:"foreignKey:CityID"`
}
